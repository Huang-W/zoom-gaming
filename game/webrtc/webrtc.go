package webrtc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v3"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
	zutils "zoomgaming/utils"
	zws "zoomgaming/websocket"
)

/**

This file can be used to negotiate a WebRTC connection.
A websocket connection is used to exchange signaling messages with the browser

The local session description is offered only once.

// Created data Channels and supported message types
// Data channels ARE negotiated in advance - make sure to create them in browser.
GameInput: pb.GameInput

// Media Tracks
None

Unable to implement Perfect negotiation w/ "impolite" peer
- Pion/webrtc does not support rollback as an SDPType
// https://w3c.github.io/webrtc-pc/#perfect-negotiation-example

*/

type WebRTC interface {
	DataChannels() chan (<-chan proto.Message)
	// Broadcast() chan (<-chan *webrtc.TrackLocalStaticRTP)
	// AddTrack(*webrtc.TrackLocalStaticRTP)
	Send(proto.Message) error // send a message to the client
	Close() error             // close the connection
}

// The server in a client-server connection between two webrtc agents
type webRTC struct {
	conn       *webrtc.PeerConnection
	videoTrack *webrtc.TrackLocalStaticRTP
	audioTrack *webrtc.TrackLocalStaticRTP
	id         uuid.UUID // identifier to distinguish this connection from others

	ws      zws.WebSocket               // WebSocket connection used for signaling
	updates chan (<-chan proto.Message) // notify the listener of any new data chhanels
	// trackUpdates      chan (<-chan *webrtc.TrackLocalStaticRTP) // notify the listener of any new media tracks from the browser
	dataChannels      map[DataChannelLabel](DataChannel) // use this mapping to send messages to the browser
	mu                *sync.Mutex                        // protects candidates
	pendingCandidates []*webrtc.ICECandidate             // save candidates for after the browser answers
}

// Constructor
func NewWebRTC(ws zws.WebSocket, videoTrack *webrtc.TrackLocalStaticRTP, audioTrack *webrtc.TrackLocalStaticRTP) (WebRTC WebRTC, err error) {

	// Catch any panics and return (nil, err) after recovering from panic
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	w := &webRTC{
		ws:         ws,
		videoTrack: videoTrack,
		audioTrack: audioTrack,
		id:         uuid.New(),
		updates:    make(chan (<-chan proto.Message)),
		// trackUpdates:      make(chan (<-chan *webrtc.TrackLocalStaticRTP)),
		dataChannels:      make(map[DataChannelLabel](DataChannel)),
		mu:                &sync.Mutex{},
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	// go routine to handle received websocket messages
	// also tears down RTCPeerConnection on death of websocket connection
	go w.watchWebSocket()

	WebRTC = w
	return
}

// Incoming datachannel... channels, and messages from those channels
func (w *webRTC) DataChannels() chan (<-chan proto.Message) {
	return w.updates
}

/**
func (w *webRTC) Broadcast() chan (<-chan *webrtc.TrackLocalStaticRTP) {
	return w.trackUpdates
}
*/
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
		log.Println("ws closing...")
		close(w.updates)
		if w.conn != nil {
			w.conn.Close()
		}
	}()

	wsOpen := w.ws.Updates()

	for {
		select {
		case ch, ok := <-wsOpen:
			if !ok {
				return
			}
			log.Println("opened!")
			for b := range ch {

				log.Println("message on ws: ")

				var msg pb.SessionDescription
				if err := proto.Unmarshal(b, msg.ProtoReflect().Interface()); err != nil {
					log.Printf("Error unmarshaling message: %v", b)
					continue
				} else {
					err := w.handleSessionDescription(&msg)
					zutils.WarnOnError(err, "Error handling client offer: %s")
				}
			}
			log.Println("closed?")
		}
	}
	log.Println("exited?")
}

// Received an offer from the browser client
//
// FIRST
func (w *webRTC) handleSessionDescription(msg *pb.SessionDescription) error {

	if msg.GetType() != pb.SessionDescription_SDP_TYPE_OFFER {
		return errors.New("expected an offer")
	}
	offerStr := msg.GetSdp()
	// log.Println(offerStr)

	// Create a MediaEngine object to configure the supported codec
	m := &webrtc.MediaEngine{}

	videoRTCPFeedback := []webrtc.RTCPFeedback{{"goog-remb", ""}, {"ccm", "fir"}, {"nack", ""}, {"nack", "pli"}}
	/**
	videoRTPCodecParameters := []webrtc.RTPCodecParameters{
		{RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, RTCPFeedback: videoRTCPFeedback}, PayloadType: 96},
	}
	*/
	// Setup the codecs you want to use.
	// We'll use a VP8 and Opus but you can also define your own
	videoRTPCodecParameters := []webrtc.RTPCodecParameters{
		{RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264, ClockRate: 90000, RTCPFeedback: videoRTCPFeedback,
			//SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f",
		}, PayloadType: 102},
		{RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264, ClockRate: 90000, RTCPFeedback: videoRTCPFeedback,
			SDPFmtpLine: "level-asymmetry-allowed=1;profile-level-id=42e01f",
		}, PayloadType: 108},
		{RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264, ClockRate: 90000, RTCPFeedback: videoRTCPFeedback,
		}, PayloadType: 123},
		{RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264, ClockRate: 90000, RTCPFeedback: videoRTCPFeedback,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f",
		}, PayloadType: 125},
		{RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType: webrtc.MimeTypeH264, ClockRate: 90000, RTCPFeedback: videoRTCPFeedback,
			SDPFmtpLine: "level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f",
		}, PayloadType: 127},
	}

	for _, codec := range videoRTPCodecParameters {
		if err := m.RegisterCodec(codec, webrtc.RTPCodecTypeVideo); err != nil {
			return err
		}
	}

	if err := m.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return err
	}

	// Create a InterceptorRegistry. This is the user configurable RTP/RTCP Pipeline.
	// This provides NACKs, RTCP Reports and other features. If you use `webrtc.NewPeerConnection`
	// this is enabled by default. If you are manually managing You MUST create a InterceptorRegistry
	// for each PeerConnection.
	i := &interceptor.Registry{}

	// Use the default set of Interceptors
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return err
	}

	// Create a setting engine. This allows influencing behavior in ways that are not support by the WebRTC API.
	e := &webrtc.SettingEngine{}

	e.SetEphemeralUDPPortRange(30000, 40000)

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i), webrtc.WithSettingEngine(*e))

	conn, err := api.NewPeerConnection(defaultRTCConfiguration)
	if err != nil {
		return err
	}
	w.conn = conn
	/**
	_, err = conn.AddTrack(w.videoTrack)
	if err != nil {
		return err
	}

	_, err = conn.AddTrack(w.audioTrack)
	if err != nil {
		return err
	}
	*/
	videoSender, err := conn.AddTrack(w.videoTrack)
	if err != nil {
		return err
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := videoSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	audioSender, err := conn.AddTrack(w.audioTrack)
	if err != nil {
		return err
	}

	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := audioSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	// WebRTC Data Channel - GameInput
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	input_impl, err := w.conn.CreateDataChannel(GameInput.String(), dcConfigs[GameInput])
	if err != nil {
		return err
	}

	input := NewDataChannel(GameInput, input_impl)
	w.dataChannels[GameInput] = input

	go func() {
		updates := input.Updates()
		for ch := range updates {
			w.updates <- ch
		}
	}()

	if err := w.conn.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offerStr,
	}); err != nil {
		return err
	}

	answer, err := w.conn.CreateAnswer(nil)
	if err != nil {
		return err
	}

	w.conn.SetLocalDescription(answer)

	w.conn.OnICECandidate(func(c *webrtc.ICECandidate) {
		// log.Println("OnIceCandidate:", c)
		if c == nil {

			local_sdp := w.conn.LocalDescription()

			// log.Println(local_sdp)

			sdp := pb.SessionDescription{}

			// pion/webrtc struct in json format
			temp_b, err := json.Marshal(local_sdp)
			if err != nil {
				log.Printf("%s", err)
				return
			}

			if err := protojson.Unmarshal(temp_b, sdp.ProtoReflect().Interface()); err != nil {
				log.Printf("%s", err)
				return
			}

			b_arr, err := proto.Marshal(sdp.ProtoReflect().Interface())
			if err != nil {
				log.Printf("%s", err)
				return
			}

			err = w.ws.Send(b_arr)
			zutils.WarnOnError(err, "Error sending sdp to browser client: %s")
		}
	})

	w.conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("ICE Connection State has changed: %s", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			log.Println("ConnectionStateConnected")
		}
		if connectionState == webrtc.ICEConnectionStateFailed || connectionState == webrtc.ICEConnectionStateClosed || connectionState == webrtc.ICEConnectionStateDisconnected {
			w.Close()
		}
	})

	w.conn.OnICEGatheringStateChange(func(s webrtc.ICEGathererState) {
		log.Println("ICE Gatherer State: ", s.String())
	})
	/**
	w.conn.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		// Create a local track, all our SFU clients will be fed via this track
		localTrack, newTrackErr := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			fmt.Sprintf("%s-%s", remoteTrack.Kind().String(), w.id.String()),
			"PlayerStream",
		)
		if newTrackErr != nil {
			log.Printf("Error creating local track: %s", newTrackErr)
			return
		}

		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
		go func() {
			ticker := time.NewTicker(rtcpPLIInterval)
			for range ticker.C {
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(remoteTrack.SSRC())}}); rtcpSendErr != nil {
					fmt.Println(rtcpSendErr)
				}
			}
		}()

		w.trackUpdates <- localTrack

		go func() {
			rtpBuf := make([]byte, 1600)
			for {
				i, _, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					log.Printf("read error: %s", err)
					return
				}

				// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
				if _, err = localTrack.Write(rtpBuf[:i]); err != nil && !errors.Is(err, io.ErrClosedPipe) {
					log.Printf("write error: %s", err)
					return
				}
			}
		}()
	})
	*/
	return nil
}
