package webrtc

import (
	"errors"
	"fmt"
	"log"
	"sync"

	_ "github.com/pion/rtp"
	webrtc "github.com/pion/webrtc/v3"

	protojson "google.golang.org/protobuf/encoding/protojson"
	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
	zutils "zoomgaming/utils"
	zws "zoomgaming/websocket"
)

/**

This file can be used to negotiate a WebRTC connection.
The connection is established using a websocket connection to signal serialized protobuf messages of type pb.SignalingEvent

The local session description is offered only once.
Any new remote session descriptions from the browser are automatically accepted and will replace any current remote sdp.

// Created data Channels and supported message types
Echo: pb.Echo

// Media Tracks
None

Current implementation does not confirm whether a SignalingEvent has been received by the client.
It does contain some out-of-order handling (needs more testing)
// https://blog.golang.org/context

*/

const (
	bufferedAmountLowThreshold uint64 = 512 * 1024  // 512 KB
	maxBufferedAmount          uint64 = 1024 * 1024 // 1 MB
)

type WebRTC interface {
	// data channel
	DataChannel(DataChannelLabel) <-chan proto.Message
	Send(DataChannelLabel, proto.Message) error
	// AttachVideoSender(<-chan *rtp.Packet) error
	Close()
}

type webRTC struct {
	conn *webrtc.PeerConnection

	ws zws.WebSocket
	// Incoming datachannel messages
	receivers         map[DataChannelLabel](chan proto.Message)
	dataChannels      map[DataChannelLabel](*webrtc.DataChannel)
	mu                *sync.Mutex // protects candidates
	pendingCandidates []*webrtc.ICECandidate
	send              chan struct{}
}

type DataChannelLabel int // Represents a unique data chhanel

const (
	Echo DataChannelLabel = iota + 1
	// GameInput
	// ChatRoom
)

func (label DataChannelLabel) String() string {
	return [...]string{"", "Echo", "GameInput", "ChatRoom"}[label]
}

var DefaultConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

// Constructor
func NewWebRTC(ws zws.WebSocket) (w WebRTC, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	webRTC := &webRTC{
		ws:                ws,
		receivers:         make(map[DataChannelLabel](chan proto.Message)),
		dataChannels:      make(map[DataChannelLabel](*webrtc.DataChannel)),
		mu:                &sync.Mutex{},
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
		send:              make(chan struct{}),
	}

	var rtcConn *webrtc.PeerConnection
	rtcConn, err = webrtc.NewPeerConnection(DefaultConfig)
	zutils.FailOnError(err, "Error setting local description: ")

	// WebRTC ICE Signaling
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// State of the WebRTC connection
	rtcConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Println("ConnectionStateConnected")
		}
		if connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed || connectionState == webrtc.ICEConnectionStateDisconnected {
			if rtcConn != nil {
				rtcConn.Close()
			}
		}
	})
	// Notify remote browser client of local server's new ICE candidate
	//
	// https://pkg.go.dev/github.com/pion/webrtc/v3#PeerConnection.OnICECandidate
	rtcConn.OnICECandidate(func(iceCandidate *webrtc.ICECandidate) {
		if iceCandidate != nil {
			webRTC.mu.Lock()
			defer webRTC.mu.Unlock()

			// Check if remote client is reading for signaling
			desc := rtcConn.RemoteDescription()
			if desc == nil {
				// If remote client is not ready, queue pending candidates
				webRTC.pendingCandidates = append(webRTC.pendingCandidates, iceCandidate)
			} else {
				webRTC.signalCandidate(iceCandidate)
			}
		}
	})
	rtcConn.OnICEGatheringStateChange(func(s webrtc.ICEGathererState) {
		log.Println("ICE Gatherer State: ", s.String())
	})
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// WebRTC Data Channel
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Keyboard events from browser come from here.
	//
	// https://pkg.go.dev/github.com/pion/webrtc#RTCDataChannel
	var dc *webrtc.DataChannel
	dc, err = rtcConn.CreateDataChannel(Echo.String(), nil)
	zutils.FailOnError(err, "Error setting local description: ")
	webRTC.dataChannels[Echo] = dc
	webRTC.receivers[Echo] = make(chan proto.Message, 1024)

	dc.OnOpen(func() {
		log.Printf("Data Channel '%s'-'%d' open.\n", dc.Label(), dc.ID())
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		b := msg.Data
		log.Printf("Message from DataChannel '%s': '%s'\n", dc.Label(), b)

		// CHANGE THIS LATER
		//
		//
		//
		// MESSAGE SHOULD BE OF TYPE PB.INPUT_EVENT, NOT PB.SIGNALING_EVENT
		var pb_msg pb.Echo
		// use name of "ierr" because NewWebRTC return variable is named "err"
		ierr := protojson.Unmarshal(b, pb_msg.ProtoReflect().Interface())
		if ierr != nil {
			return
		}

		// Send across receiving go channel
		webRTC.receivers[Echo] <- pb_msg.ProtoReflect().Interface()
	})
	// Set bufferedAmountLowThreshold so that we can get notified when
	// we can send more
	dc.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)
	// This callback is made when the current bufferedAmount becomes lower than the threadshold
	dc.OnBufferedAmountLow(func() {
		webRTC.send <- struct{}{}
	})
	dc.OnClose(func() {
		log.Println("Data channel closed")
		close(webRTC.receivers[Echo])
		log.Println("Closed webrtc")
	})
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// Send STUN server used in this server's config to the browser client
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	var iceServer webrtc.ICEServer
	iceServer = DefaultConfig.ICEServers[0]

	var iceServer_pb pb.RTCIceServer
	err = zutils.ConvertToProtoMessage(&iceServer, iceServer_pb.ProtoReflect())
	zutils.FailOnError(err, "Error converting ice server: ")

	evt1 := pb.SignalingEvent{
		Event: &pb.SignalingEvent_RtcIceServer{
			RtcIceServer: &iceServer_pb,
		},
	}

	err = ws.Send(evt1.ProtoReflect().Interface())
	zutils.FailOnError(err, "Error sending iceServer to browser client: ")
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	// Send this server's local sdp to the browser client
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	var localSdp webrtc.SessionDescription
	localSdp, err = rtcConn.CreateOffer(nil)
	zutils.FailOnError(err, "Error setting local description: ")

	var localSdp_pb pb.SessionDescription
	err = zutils.ConvertToProtoMessage(&localSdp, localSdp_pb.ProtoReflect())
	zutils.FailOnError(err, "Error converting local description: ")

	evt2 := pb.SignalingEvent{
		Event: &pb.SignalingEvent_SessionDescription{
			SessionDescription: &localSdp_pb,
		},
	}

	err = ws.Send(evt2.ProtoReflect().Interface())
	zutils.FailOnError(err, "Error sending sdp to browser client: ")

	err = rtcConn.SetLocalDescription(localSdp)
	zutils.FailOnError(err, "Error setting local description: ")
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	webRTC.conn = rtcConn
	go webRTC.signalingEvents()

	w = webRTC
	return
}

// Incoming datachannel messages
func (w *webRTC) DataChannel(label DataChannelLabel) <-chan proto.Message {
	return w.receivers[label]
}

// Send a message to the client
func (w *webRTC) Send(label DataChannelLabel, msg proto.Message) error {
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	dc, prs := w.dataChannels[label]
	if !prs {
		return errors.New("Label not found")
	}

	if dc.BufferedAmount()+uint64(len(b)) > maxBufferedAmount {
		<-w.send
	}

	err = dc.Send(b)

	return err
}

func (w *webRTC) Close() {
	_ = w.conn.Close()
	_ = w.ws.Close()
}

// Handle any received websocket messages
//
// Supported Events:
//	pb.SessionDescription
//	pb.RtcIceCandidateInit
//
// Unhandled Events:
//	pb.RtcIceServer
func (w *webRTC) signalingEvents() {
	clientUpdates := w.ws.Updates()
	for {
		select {
		case msg, ok := <-clientUpdates:
			if !ok {
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

	var cand pb.RTCIceCandidateInit
	err := zutils.ConvertToProtoMessage(&iceCandInit, cand.ProtoReflect())
	if err != nil {
		log.Println("Error converting proto message: ", err)
		return
	}

	msg := pb.SignalingEvent{
		Event: &pb.SignalingEvent_RtcIceCandidateInit{
			RtcIceCandidateInit: &cand,
		},
	}

	err = w.ws.Send(msg.ProtoReflect().Interface())
	zutils.WarnOnError(err, "Error sending ice candidate: ")
}
