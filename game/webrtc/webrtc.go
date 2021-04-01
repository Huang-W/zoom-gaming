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

// https://blog.golang.org/context

*/

type WebRTC interface {
	DataChannels() chan (<-chan proto.Message)
	Send(proto.Message) error // send a message to the client
	// AttachVideoSender(<-chan *rtp.Packet) error
	Close() error // close the connection
}

// The server in a client-server connection between two webrtc agents
type webRTC struct {
	conn *webrtc.PeerConnection

	ws                zws.WebSocket                      // WebSocket connection used for signaling
	updates           chan (<-chan proto.Message)        // notify the listener of any new data chhanels
	dataChannels      map[DataChannelLabel](DataChannel) // use this mapping to send messages to the browser
	mu                *sync.Mutex                        // protects candidates
	pendingCandidates []*webrtc.ICECandidate             // save candidates for after the browser answers
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
		updates:           make(chan (<-chan proto.Message)),
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
	// also tears down RTCPeerConnection on death of websocket connection
	go w.watchWebSocket()

	WebRTC = w
	return
}

// Incoming datachannel... channels, and messages from those channels
func (w *webRTC) DataChannels() chan (<-chan proto.Message) {
	return w.updates
}

// Send a message to the browser idiomatically based on message type
func (w *webRTC) Send(msg proto.Message) error {
	label, prs := reverseMapping[msg.ProtoReflect().Type()]
	if !prs {
		return errors.New("Invalid message type")
	}
	dc, prs := w.dataChannels[label]
	if !prs {
		return errors.New(fmt.Sprintf("Data Channel with label %s not found", label))
	}
	return dc.Send(msg)
}

func (w *webRTC) Close() error {
	return w.ws.Close() // close the websocket connection
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

	echo := NewDataChannel(Echo, echo_impl)

	w.dataChannels[Echo] = echo
	go w.onDataChannelOpen(echo)
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// WebRTC Data Channel - GameInput
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Keyboard events from browser come from here.
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
}

func (w *webRTC) onDataChannelOpen(dc DataChannel) {
	updates := dc.Updates()
	for {
		select {
		case ch, ok := <-updates:
			if !ok {
				return
			}
			w.updates <- ch
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

// Handle any websocket messages received from the client
//
// Expected Events:
//	pb.SessionDescription
//	pb.RtcIceCandidateInit
//
// Ignored Events:
//	pb.RtcIceServer
func (w *webRTC) watchWebSocket() {

	defer func() {
		// Start the teardown sequence and close all data channels
		for _, dc := range w.dataChannels {
			dc.Close()
		}
		w.conn.Close()
	}()

	updates := w.ws.Updates()

	for {
		select {
		case ch, ok := <-updates:

			if !ok {
				return
			}

			go w.handleSignalingEvents(ch)

		}
	}
}

func (w *webRTC) handleSignalingEvents(ch <-chan []byte) {
	for {
		select {
		case b, ok := <-ch:
			if !ok {
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

				if err := w.handleSessionDescription(&msg); err != nil {
					zutils.WarnOnError(err, "Error handling session description message on WS: ")
				}

			case *pb.SignalingEvent_RtcIceCandidateInit:

				if err := w.handleRtcIceCandidateInit(&msg); err != nil {
					zutils.WarnOnError(err, "Error handling ice cand message on WS: ")
				}

			case nil:

				log.Println("Signaling event message with empty Event field")

			default:

				log.Printf("SignalingEvent.Event has unexpected type %T", t)

			}
		}
	}
}

func (w *webRTC) handleSessionDescription(msg *pb.SignalingEvent) error {
	sdp_pb := msg.GetSessionDescription()

	var sdp_pion webrtc.SessionDescription
	if err := zutils.ConvertFromProtoMessage(sdp_pb.ProtoReflect(), &sdp_pion); err != nil {
		return err
	}

	if err := w.conn.SetRemoteDescription(sdp_pion); err != nil {
		return err
	}

	w.mu.Lock()
	func() {
		defer w.mu.Unlock()
	}()

	for _, c := range w.pendingCandidates {
		w.signalCandidate(c)
	}

	return nil
}

func (w *webRTC) handleRtcIceCandidateInit(msg *pb.SignalingEvent) error {
	rtcIceCand_pb := msg.GetRtcIceCandidateInit()

	var rtcIceCand_pion webrtc.ICECandidateInit
	if err := zutils.ConvertFromProtoMessage(rtcIceCand_pb.ProtoReflect(), &rtcIceCand_pion); err != nil {
		return err
	}

	return w.conn.AddICECandidate(rtcIceCand_pion)
}
