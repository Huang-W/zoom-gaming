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
The only track/channel the browser is allowed to add are the webcam audio + video tracks

// Created data Channels and supported message types
// Data channels are NOT negotiated - make sure to create them in browser.
Echo: pb.Echo

// Media Tracks
None

This code does not handle disconnects and probably a lot of other stuff.
// https://blog.golang.org/context

*/

type WebRTC interface {
	DataChannel(DataChannelLabel) <-chan proto.Message // a stream of data channel messages from the browser
	Send(DataChannelLabel, proto.Message) error        // send a message to the client
	// AttachVideoSender(<-chan *rtp.Packet) error
	Close() error // close the connection
}

// The server in a client-server connection between two webrtc agents
type webRTC struct {
	conn *webrtc.PeerConnection

	ws                zws.WebSocket                              // WebSocket connection used for signaling
	receivers         map[DataChannelLabel](chan proto.Message)  // data channel messages arrive here
	dataChannels      map[DataChannelLabel](*webrtc.DataChannel) // internal handling of data channels
	flowControls      map[DataChannelLabel](chan struct{})       // prevent overflowing of data channel send buffer
	mu                *sync.Mutex                                // protects candidates
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
		receivers:         make(map[DataChannelLabel](chan proto.Message)),
		dataChannels:      make(map[DataChannelLabel](*webrtc.DataChannel)),
		flowControls:      make(map[DataChannelLabel](chan struct{})),
		mu:                &sync.Mutex{},
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	var (
		// pion/webrtc
		conn      *webrtc.PeerConnection
		iceServer webrtc.ICEServer
		localSdp  webrtc.SessionDescription
	)

	conn, err = webrtc.NewPeerConnection(defaultConfig)
	zutils.FailOnError(err, "Error setting local description: ")

	w.conn = conn
	w.initPeerConnectionHandlers() // register RTCPeerConnection handlers
	w.initDataChannels()           // Create pre-negotiated data channels and register data channel handlers

	// The STUN server used in this server's configuration
	iceServer = defaultConfig.ICEServers[0]

	// wrap the ice server as a pb.SignalingEvent message
	iceServer_msg, err := zutils.WrapRTCIceServer(&iceServer)
	zutils.FailOnError(err, "Error converting ice server: ")

	// Send the message to the client
	err = ws.Send(iceServer_msg.ProtoReflect().Interface())
	zutils.FailOnError(err, "Error sending iceServer to browser client: ")

	// Create an offer
	localSdp, err = conn.CreateOffer(nil)
	zutils.FailOnError(err, "Error creating offer: ")

	// wrap the session description as a pb.SignalingEvent message
	localSdp_msg, err := zutils.WrapSessionDescription(&localSdp)
	zutils.FailOnError(err, "Error converting local description: ")

	// Send this server's session description to the browser client
	err = ws.Send(localSdp_msg.ProtoReflect().Interface())
	zutils.FailOnError(err, "Error sending sdp to browser client: ")

	err = conn.SetLocalDescription(localSdp)
	zutils.FailOnError(err, "Error setting local description: ")
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// go routine to handle received websocket messages
	//
	// is also used to close the webrtc connection
	go w.signalingEvents()

	WebRTC = w
	return
}

// Incoming datachannel messages
func (w *webRTC) DataChannel(label DataChannelLabel) <-chan proto.Message {
	return w.receivers[label]
}

// Send a message to the client
func (w *webRTC) Send(label DataChannelLabel, msg proto.Message) error {

	// marshal the message into a byte array
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	// check if the data channel has been created
	dc, prs := w.dataChannels[label]
	if !prs {
		return errors.New("Label not found")
	}

	// block if the size of the stream's buffer exceeds 1 MB
	if dc.BufferedAmount()+uint64(len(b)) > maxBufferedAmount {
		<-w.flowControls[label]
	}

	err = dc.Send(b)

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
		case msg, ok := <-clientUpdates:
			if !ok {
				// Start the clean-up sequence

				// close all data channels
				for _, dc := range w.dataChannels {
					dc.Close()
				}

				// close receiving go channels
				for _, ch := range w.receivers {
					close(ch)
				}

				// close buffering controls
				for _, ch := range w.flowControls {
					close(ch)
				}

				w.conn.Close()
				return
			}

			if msg_pb, ok := msg.(*pb.SignalingEvent); ok {

				evt := msg_pb.GetEvent()

				switch t := evt.(type) {

				case *pb.SignalingEvent_RtcIceServer:
					continue // not expected from client

				case *pb.SignalingEvent_SessionDescription:
					sdp_pb := msg_pb.GetSessionDescription()

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
					rtcIceCand_pb := msg_pb.GetRtcIceCandidateInit()

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
			} else {
				log.Println(msg)
			}
		}
	}
}

// Helper function to marshal and send local ice candidates
func (w *webRTC) signalCandidate(c *webrtc.ICECandidate) {
	iceCandInit := c.ToJSON()
	log.Println("OnIceCandidate:", iceCandInit.Candidate)

	msg, err := zutils.WrapRTCIceCandidateInit(&iceCandInit)
	zutils.FailOnError(err, "Error converting iceCandInit: ")

	err = w.ws.Send(msg.ProtoReflect().Interface())
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
	var dc *webrtc.DataChannel
	dc, err := w.conn.CreateDataChannel(Echo.String(), &webrtc.DataChannelInit{
		Ordered:    &echo_ordered,
		Negotiated: &echo_negotiated,
		ID:         &echo_id,
	})
	zutils.FailOnError(err, "Error setting local description: ")

	w.dataChannels[Echo] = dc
	w.receivers[Echo] = make(chan proto.Message, 1024)
	w.flowControls[Echo] = make(chan struct{})

	dc.OnOpen(func() {
		log.Printf("Data Channel '%s'-'%d' open.\n", dc.Label(), dc.ID())
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		b := msg.Data
		log.Printf("Message from DataChannel '%s': '%s'\n", dc.Label(), b)

		var pb_msg pb.Echo
		err := proto.Unmarshal(b, pb_msg.ProtoReflect().Interface())
		if err != nil {
			return
		}

		// Send across receiving go channel
		w.receivers[Echo] <- pb_msg.ProtoReflect().Interface()
	})
	// Set bufferedAmountLowThreshold so that we can get notified when
	// we can send more
	dc.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)
	// This callback is made when the current bufferedAmount becomes lower than the threadshold
	dc.OnBufferedAmountLow(func() {
		w.flowControls[Echo] <- struct{}{}
	})
	dc.OnClose(func() {
		log.Println("Data channel closed")
	})
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// WebRTC Data Channel - GameInput
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Keyboard events from browser come from here.
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
}
