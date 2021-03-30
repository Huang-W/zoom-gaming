package webrtc

import (
	"errors"
	"fmt"
	"log"
	"sync"

	_ "github.com/pion/rtp"
	webrtc "github.com/pion/webrtc/v3"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
	zutils "zoomgaming/utils"
	zws "zoomgaming/websocket"
)

/**

This file can be used to negotiate a WebRTC connection.
A websocket connection is used to exchange signaling messages with the browser

The local session description is offered only once.

// Created data Channels and supported message types
// Data channels are NOT negotiated - make sure to create them in browser.
Echo: pb.Echo

// Media Tracks
None

This code does not handle disconnects and probably a lot of other stuff.
// https://blog.golang.org/context

*/

type WebRTC interface {
	DataChannel(DataChannelLabel) (<-chan proto.Message, error) // a stream of data channel messages from the browser
	Send(DataChannelLabel, proto.Message) error                 // send a message to the client
	// AttachVideoSender(<-chan *rtp.Packet) error
	Close() error // close the connection
}

// The server in a client-server connection between two webrtc agents
type webRTC struct {
	conn *webrtc.PeerConnection

	ws                zws.WebSocket                      // WebSocket connection used for signaling
	dataChannels      map[DataChannelLabel](DataChannel) // data channels by label
	mu                *sync.Mutex                        // protects candidates
	pendingCandidates []*webrtc.ICECandidate
}

// Constructor
func NewWebRTC(ws zws.WebSocket) (WebRTC WebRTC, err error) {

	// Catch any panics and return (nil, err) after recovering from panic
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	w := &webRTC{
		ws:                ws,
		dataChannels:      make(map[DataChannelLabel](DataChannel)),
		mu:                &sync.Mutex{},
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	var (
		// pion/webrtc
		conn      *webrtc.PeerConnection
		iceServer webrtc.ICEServer
		sdp       webrtc.SessionDescription
	)

	conn, err = webrtc.NewPeerConnection(defaultRTCConfiguration)
	zutils.FailOnError(err, "Error setting local description: ")

	w.conn = conn
	w.initPeerConnectionHandlers() // register RTCPeerConnection handlers
	w.initDataChannels()           // Create pre-negotiated data channels and register data channel handlers

	// Send the STUN server used in this server's configuration
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	iceServer = defaultRTCConfiguration.ICEServers[0]

	b1, err := zutils.MarshalSignalingEvent(&iceServer)
	zutils.FailOnError(err, "Error converting ice server: ")

	err = ws.Send(b1)
	zutils.FailOnError(err, "Error sending iceServer to browser client: ")
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// Create an offer and send to browser
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	sdp, err = conn.CreateOffer(nil)
	zutils.FailOnError(err, "Error creating offer: ")

	b2, err := zutils.MarshalSignalingEvent(&sdp)
	zutils.FailOnError(err, "Error converting local description: ")

	err = ws.Send(b2)
	zutils.FailOnError(err, "Error sending sdp to browser client: ")

	err = conn.SetLocalDescription(sdp)
	zutils.FailOnError(err, "Error setting local description: ")
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// go routine to handle received websocket messages
	//
	// also handles teardown
	go w.signalingEvents()

	WebRTC = w
	return
}

// Incoming datachannel messages
func (w *webRTC) DataChannel(label DataChannelLabel) (<-chan proto.Message, error) {
	dc, prs := w.dataChannels[label]
	if !prs {
		return nil, errors.New(fmt.Sprintf("Channel with label %s not found", label))
	}
	return dc.Updates(), nil
}

// Send a message to the browser on a specific data channel
func (w *webRTC) Send(label DataChannelLabel, msg proto.Message) error {
	dc, prs := w.dataChannels[label]
	if !prs {
		return errors.New(fmt.Sprintf("Data channel with label of %s not found", label))
	}

	err := dc.Send(msg)
	return err
}

func (w *webRTC) Close() (err error) {
	err = w.ws.Close() // close the websocket connection
	return
}

// Handle any websocket messages received from the client
//
// Expected Events:
//	pb.SessionDescription
//	pb.RtcIceCandidateInit
//
// Ignored Events:
//	pb.RtcIceServer
func (w *webRTC) signalingEvents() {
	clientUpdates := w.ws.Updates()
	for {
		select {
		case b, ok := <-clientUpdates:
			if !ok {
				// Start the teardown sequence

				// close all data channels
				for _, dc := range w.dataChannels {
					dc.Close()
				}

				w.conn.Close()
				return
			}

			var msg pb.SignalingEvent
			if err := proto.Unmarshal(b, msg.ProtoReflect().Interface()); err != nil {
				log.Printf("Error unmarshaling message: %v", b)
				continue
			}

			evt := msg.GetEvent()

			switch t := evt.(type) {

			case *pb.SignalingEvent_RtcIceServer:
				continue // not expected from client

			case *pb.SignalingEvent_SessionDescription:
				sdp_pb := msg.GetSessionDescription()

				var sdp_pion webrtc.SessionDescription
				err := zutils.ConvertFromProtoMessage(sdp_pb.ProtoReflect(), &sdp_pion)
				if err != nil {
					log.Println(err)
					continue
				}

				err = w.conn.SetRemoteDescription(sdp_pion)
				if err != nil {
					log.Println(err)
					continue
				}

				w.mu.Lock()
				func() {
					defer w.mu.Unlock()
				}()

				for _, c := range w.pendingCandidates {
					w.signalCandidate(c)
				}

			case *pb.SignalingEvent_RtcIceCandidateInit:
				rtcIceCand_pb := msg.GetRtcIceCandidateInit()

				var rtcIceCand_pion webrtc.ICECandidateInit
				err := zutils.ConvertFromProtoMessage(rtcIceCand_pb.ProtoReflect(), &rtcIceCand_pion)
				if err != nil {
					log.Println(err)
					continue
				}

				err = w.conn.AddICECandidate(rtcIceCand_pion)
				zutils.WarnOnError(err, "Error adding ICE Candidate: ")

			case nil:
				log.Println("Signaling event message with empty Event field")
			default:
				log.Printf("SignalingEvent.Event has unexpected type %T", t)
			}
		}
	}
}

// Helper function to marshal and send local ice candidates
func (w *webRTC) signalCandidate(c *webrtc.ICECandidate) {
	iceCandInit := c.ToJSON()
	log.Println("OnIceCandidate:", iceCandInit.Candidate)

	b, err := zutils.MarshalSignalingEvent(&iceCandInit)
	zutils.FailOnError(err, "Error converting iceCandInit: ")

	err = w.ws.Send(b)
	zutils.WarnOnError(err, "Error sending ice candidate: ")
}

func (w *webRTC) initPeerConnectionHandlers() {

	w.conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Println("ConnectionStateConnected")
		}
		if connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed || connectionState == webrtc.ICEConnectionStateDisconnected {
			if w.conn != nil {
				w.conn.Close()
			}
		}
	})

	// An ICE Candidate represents one end of the underlying RTCDtlsTransport
	//
	// A pair of ICE candidates, one from each agent, establishes a connection
	//
	// The local server is the controlling ICE agent
	//
	// https://pkg.go.dev/github.com/pion/webrtc/v3#PeerConnection.OnICECandidate
	w.conn.OnICECandidate(func(iceCandidate *webrtc.ICECandidate) {

		if iceCandidate != nil {

			w.mu.Lock()
			defer w.mu.Unlock()

			desc := w.conn.RemoteDescription()
			if desc == nil {

				// If remote description has not been set, queue pending candidates
				w.pendingCandidates = append(w.pendingCandidates, iceCandidate)

			} else {

				// Else, send the candidate to the client
				w.signalCandidate(iceCandidate)

			}
		}
	})
	w.conn.OnICEGatheringStateChange(func(s webrtc.ICEGathererState) {
		log.Println("ICE Gatherer State: ", s.String())
	})
}

func (w *webRTC) initDataChannels() {

	// WebRTC Data Channel - Echo
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	echo_impl, err := w.conn.CreateDataChannel(Echo.String(), dcConfigs[Echo])
	zutils.FailOnError(err, "Error creating data channel: ")

	echo, err := NewDataChannel(Echo, echo_impl)
	zutils.FailOnError(err, "Error creating data channel interface: ")

	w.dataChannels[Echo] = echo
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// WebRTC Data Channel - GameInput
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Keyboard events from browser come from here.
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
}
