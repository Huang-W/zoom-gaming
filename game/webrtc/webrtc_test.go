package webrtc

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	websocket "github.com/gorilla/websocket"
	webrtc "github.com/pion/webrtc/v3"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
	zutils "zoomgaming/utils"
	zws "zoomgaming/websocket"
)

var upgrader = websocket.Upgrader{}

// Represents the server in a client-server connection between two webrtc agents
func dataChannelEcho(w http.ResponseWriter, r *http.Request) {

	// HTTP Upgrade to websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Start a new webrtc agent using the websocket connection
	ws := zws.NewWebSocket(c)
	rtc, err := NewWebRTC(ws)
	if err != nil {
		log.Println(err)
		return
	}

	// Newly opened data channels arrive here
	updates := rtc.DataChannels()

	// Echo messages back to the client
	for {
		select {
		case ch, ok := <-updates:
			if !ok {
				return
			}
			// start a new go routine that echoes back Data Channel messages
			go func() {
				for {
					select {
					case msg, ok := <-ch:
						if !ok {
							return
						}
						rtc.Send(msg)
					}
				}
			}()
		}
	}
}

// Test the WebRTC server response
func TestDataChannel(t *testing.T) {

	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(dataChannelEcho))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	WS := zws.NewWebSocket(ws)

	// Rerepsents a browser client that is trying to connect to the server
	testClient := &testClient{
		t:            t,
		ws:           WS,
		dataChannels: make(map[DataChannelLabel](DataChannel)),
		updates:      make(chan (<-chan proto.Message)),
	}

	go testClient.watchWebSocket()

	// (for testing) Start a go routine that sends an empty message to the server once every second
	ticker := time.NewTicker(1 * time.Second)

	// Used to exit the "Send" go routine
	finish := make(chan struct{})

	// The "Send" go routine
	go func() {
		for {
			select {
			case <-finish:
				ticker.Stop()
				testClient.Close()
				return
			case <-ticker.C:
				test_msg := pb.Echo{}

				if err = testClient.Send(test_msg.ProtoReflect().Interface()); err != nil {
					t.Logf("Data channel has not been created yet: %s", err.Error())
				}
			}
		}
	}()

	// The "Receive" go routine
	go func() {
		updates := testClient.DataChannels()
		for {
			select {
			case ch, ok := <-updates:
				if !ok {
					return
				}
				// when a data channel has been created, start a go routine that logs messages echoed from the server
				go func() {
					for {
						select {
						case msg, ok := <-ch:
							if !ok {
								return
							}
							b, _ := proto.Marshal(msg)
							t.Logf("%s", b)
						}
					}
				}()
			}
		}
	}()

	// After 10 seconds, close the connection
	time.Sleep(10 * time.Second)
	finish <- struct{}{}
}

/**
 * The client in a client-server connection between two webrtc agents
 *
 * A Webrtc connection is established by signaling messages across a websocket connection.
 *
 * The client agent implementation differs from the server agent implementation
 */
type testClient struct {
	conn *webrtc.PeerConnection
	t    *testing.T

	ws           zws.WebSocket
	dataChannels map[DataChannelLabel](DataChannel) // internal handling of data channels
	updates      chan (<-chan proto.Message)        // wait for each data channel to be created
}

func (client *testClient) DataChannels() chan (<-chan proto.Message) {
	return client.updates
}

// Send a message across the datachannel
func (client *testClient) Send(msg proto.Message) error {
	label, prs := reverseMapping[msg.ProtoReflect().Type()]
	if !prs {
		return errors.New("Invalid message type")
	}
	dc, prs := client.dataChannels[label]
	if !prs {
		return errors.New(fmt.Sprintf("Data Channel with label %s not found", label))
	}
	return dc.Send(msg)
}

// Start a chain reaction and close the connection
func (client *testClient) Close() error {
	return client.ws.Close()
}

// go routine to process websocket messages
//
// Closes active channels when the websocket connection is CLOSED
func (client *testClient) watchWebSocket() {

	defer func() {
		// Start the teardown sequence and close all data channels
		for _, dc := range client.dataChannels {
			dc.Close()
		}
		client.conn.Close()
	}()

	updates := client.ws.Updates()

	for {
		select {
		case ch, ok := <-updates:

			if !ok {
				return
			}

			go client.handleSignalingEvents(ch)

		}
	}
}

func (client *testClient) handleSignalingEvents(ch <-chan []byte) {

	for {
		select {
		case b, ok := <-ch:

			if !ok {
				return
			}

			var msg pb.SignalingEvent
			if err := proto.Unmarshal(b, msg.ProtoReflect().Interface()); err != nil {
				client.t.Logf("Error unmarshaling message: %v", b)
				continue
			}

			evt := msg.GetEvent()

			switch evt.(type) {

			case *pb.SignalingEvent_RtcIceServer:

				if err := client.handleRtcIceServer(&msg); err != nil {
					client.t.Logf("Error handling rtc ice server msg on WS: ")
				}

			case *pb.SignalingEvent_SessionDescription:

				if err := client.handleSessionDescription(&msg); err != nil {
					client.t.Logf("Error handling session description msg on WS: ")
				}

			case *pb.SignalingEvent_RtcIceCandidateInit:

				if err := client.handleRtcIceCandidateInit(&msg); err != nil {
					client.t.Logf("Error handling ice candidate msg on WS: ")
				}

			case nil:
				client.t.Logf("Signaling event message with empty Event field")
			default:
				client.t.Logf("SignalingEvent.Event has unexpected type %T", evt)
			}
		}
	}
}

func (client *testClient) handleRtcIceServer(msg *pb.SignalingEvent) error {
	// 1. Convert from protobuf to pion/webrtc
	iceServer_pb := msg.GetRtcIceServer()

	// Convert from protobuf to pion/webrtc
	var iceServer_pion webrtc.ICEServer
	if err := zutils.ConvertFromProtoMessage(iceServer_pb.ProtoReflect(), &iceServer_pion); err != nil {
		return err
	}

	// RTCConfiguration
	var config = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{iceServer_pion},
	}

	// 2. Initialize the local webrtc agent
	conn, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}
	client.conn = conn

	// 3. Add event handlers
	client.initClientHandlers()

	// 4. Create a pre-negotiated datachannel with label of "Echo" and ID of "1111"
	client.initEchoDataChannel()

	return nil
}

func (client *testClient) handleSessionDescription(msg *pb.SignalingEvent) error {

	sdp_pb := msg.GetSessionDescription()

	// Convert from protobuf to pion/webrtc
	var sdp_pion webrtc.SessionDescription
	if err := zutils.ConvertFromProtoMessage(sdp_pb.ProtoReflect(), &sdp_pion); err != nil {
		return err
	}

	// Set remote sdp and send local sdp to server
	if err := client.conn.SetRemoteDescription(sdp_pion); err != nil {
		return err
	}

	answer, err := client.conn.CreateAnswer(nil)
	if err != nil {
		return err
	}

	if err := client.signalSessionDescription(&answer); err != nil {
		return err
	}

	return client.conn.SetLocalDescription(answer)
}

func (client *testClient) handleRtcIceCandidateInit(msg *pb.SignalingEvent) error {
	// 1. Convert from protobuf to pion/webrtc
	rtcIceCand_pb := msg.GetRtcIceCandidateInit()

	// Convert from protobuf to pion/webrtc
	var rtcIceCand_pion webrtc.ICECandidateInit
	if err := zutils.ConvertFromProtoMessage(rtcIceCand_pb.ProtoReflect(), &rtcIceCand_pion); err != nil {
		return err
	}

	// 2. Add a remote ICE candidate
	return client.conn.AddICECandidate(rtcIceCand_pion)
}

// register handlers for the RTCPeerConnection
func (client *testClient) initClientHandlers() {

	// When an ICE candidate is available send to the "server" instance
	client.conn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		if err := client.signalCandidate(c); err != nil {
			client.t.Logf("%s", err)
		}
	})

	// Set the handler for ICE connection state
	// This will notify you when the peer has connected/disconnected
	client.conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		client.t.Logf("Client ICE Connection State has changed: %s\n", connectionState.String())
	})
}

// Create a pre-negotiated data channel with label of "Echo" and ID of "1111"
//
// Register OnMessage handler
func (client *testClient) initEchoDataChannel() error {
	echo_impl, err := client.conn.CreateDataChannel(Echo.String(), dcConfigs[Echo])
	if err != nil {
		return err
	}

	echo := NewDataChannel(Echo, echo_impl)

	client.dataChannels[Echo] = echo // this is used in the Send(DataChannelLabel, proto.Message) function
	go client.onDataChannelOpen(echo)

	return nil
}

func (client *testClient) onDataChannelOpen(dc DataChannel) {
	updates := dc.Updates()
	for {
		select {
		case ch, ok := <-updates:
			if !ok {
				return
			}
			client.updates <- ch
		}
	}
}

// Helper function to send an RtcIceCandidateInit to the server
// over the websocket connection
func (client *testClient) signalCandidate(c *webrtc.ICECandidate) error {

	// Convert to dictionary form
	iceCandInit := c.ToJSON()

	// Wrap the ice candidate as a pb.SignalingEvent
	b, err := zutils.MarshalSignalingEvent(&iceCandInit)
	if err != nil {
		return err
	}

	// Send the message over websocket
	if err := client.ws.Send(b); err != nil {
		return err
	}

	return nil
}

// Helper function to send a SessionDescription to the server
// over the websocket connection
func (client *testClient) signalSessionDescription(sdp *webrtc.SessionDescription) error {

	// Wrap the sdp as a pb.SignalingEvent
	b, err := zutils.MarshalSignalingEvent(sdp)
	if err != nil {
		return err
	}

	// Send the message over websocket
	if err := client.ws.Send(b); err != nil {
		return err
	}

	return nil
}
