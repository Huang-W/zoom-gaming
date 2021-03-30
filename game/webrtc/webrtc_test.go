package webrtc

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

	var (
		c   *websocket.Conn
		ws  zws.WebSocket
		rtc WebRTC
	)

	// HTTP Upgrade to websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Start a new webrtc agent using the websocket connection
	ws = zws.NewWebSocket(c)
	rtc, err = NewWebRTC(ws)
	if err != nil {
		log.Println(err)
		return
	}

	// Client messages arrive on this channel
	ch, err := rtc.DataChannel(Echo)
	if err != nil {
		log.Println(err)
		return
	}

	// Echo messages back to the client
	for {
		select {
		case msg, ok := <-ch:
			log.Println(msg)
			if !ok {
				return
			}
			rtc.Send(Echo, msg)
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
		t:                 t,
		ws:                WS,
		dataChannels:      make(map[DataChannelLabel](DataChannel)),
		Receiver:          make(chan (<-chan proto.Message)),
		mu:                &sync.Mutex{},
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	go testClient.signalingEvents()

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

				if err = testClient.Send(Echo, test_msg.ProtoReflect().Interface()); err != nil {
					t.Logf("Data channel has not been created yet: %s", err.Error())
				}
			}
		}
	}()

	go func() {
		for {
			select {
			case ch, ok := <-testClient.Receiver:
				if !ok {
					return
				}
				// when a data channel has been created, start a go routine that logs receieved messages
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
	Receiver     chan (<-chan proto.Message)        // wait for each data channel to be created
	mu           *sync.Mutex                        // protects the pendingCandidates variable

	// After the remote session description is set, the candidates in pendingCandidates will be signaled to the remote agent
	pendingCandidates []*webrtc.ICECandidate

	/**
	 * pendingSdp holds the remote session description in case the
	 * "pb.SessionDescription" message is received before the local RTCPeerConnection is initialized
	 */
	pendingSdp *webrtc.SessionDescription
}

// Send a message across the datachannel
func (client *testClient) Send(label DataChannelLabel, msg proto.Message) error {

	// Check if the data channel has been created
	dc, prs := client.dataChannels[label]
	if !prs {
		return errors.New(fmt.Sprintf("Data channel with label of %s not found", label))
	}

	err := dc.Send(msg)
	return err
}

// Start a chain reaction and close the connection
func (client *testClient) Close() (err error) {
	// Send a websocket close event
	err = client.ws.Close()
	return
}

// go routine to process websocket messages
//
// Closes active channels when the websocket connection is CLOSED
func (client *testClient) signalingEvents() {
	clientUpdates := client.ws.Updates()
	for {
		select {
		case b, ok := <-clientUpdates:
			if !ok {
				// close all data channels
				for _, dc := range client.dataChannels {
					dc.Close()
				}

				client.conn.Close()
				close(client.Receiver)
				return
			}

			var msg pb.SignalingEvent
			if err := proto.Unmarshal(b, msg.ProtoReflect().Interface()); err != nil {
				client.t.Logf("Error unmarshaling message: %v", b)
				continue
			}

			evt := msg.GetEvent()

			switch evt.(type) {

			// Event Type 1
			// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
			case *pb.SignalingEvent_RtcIceServer:

				// 1. Convert from protobuf to pion/webrtc
				iceServer_pb := msg.GetRtcIceServer()

				// Convert from protobuf to pion/webrtc
				var iceServer_pion webrtc.ICEServer
				if err := zutils.ConvertFromProtoMessage(iceServer_pb.ProtoReflect(), &iceServer_pion); err != nil {
					client.t.Errorf("%s", err)
				}

				// RTCConfiguration
				var config = webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{iceServer_pion},
				}

				// 2. Initialize the local webrtc agent
				var err error
				client.conn, err = webrtc.NewPeerConnection(config)
				if err != nil {
					client.t.Errorf("%s", err)
				}

				// 3. Add event handlers
				client.initClientHandlers()

				// 4. Create a pre-negotiated datachannel with label of "Echo" and ID of "1111"
				client.initEchoDataChannel()

				// Protect the client.pendingCandidates variable
				client.mu.Lock()
				func() {
					defer client.mu.Unlock()
				}()

				// 5. Handle case where session description is received early
				if client.pendingSdp != nil {
					if err := client.conn.SetRemoteDescription(*client.pendingSdp); err != nil {
						client.t.Errorf("%s", err)
					}

					answer, err := client.conn.CreateAnswer(nil)
					if err != nil {
						client.t.Errorf("%s", err)
					}

					if err := client.signalSessionDescription(&answer); err != nil {
						client.t.Errorf("%s", err)
					}

					if err := client.conn.SetLocalDescription(answer); err != nil {
						client.t.Errorf("%s", err)
					}

					for _, c := range client.pendingCandidates {
						if err = client.signalCandidate(c); err != nil {
							client.t.Logf("%s", err)
						}
					}
				}

			// Event Type 2
			// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
			case *pb.SignalingEvent_SessionDescription:

				// 1. Convert from protobuf to pion/webrtc
				sdp_pb := msg.GetSessionDescription()

				// Convert from protobuf to pion/webrtc
				var sdp_pion webrtc.SessionDescription
				if err := zutils.ConvertFromProtoMessage(sdp_pb.ProtoReflect(), &sdp_pion); err != nil {
					client.t.Errorf("%s", err)
				}

				// 2. If conn == nil, session description was received too early
				if client.conn == nil {
					client.pendingSdp = &sdp_pion
					continue
				}

				// 3. Else, set remote sdp and send local sdp to server
				if err := client.conn.SetRemoteDescription(sdp_pion); err != nil {
					client.t.Errorf("%s", err)
				}

				answer, err := client.conn.CreateAnswer(nil)
				if err != nil {
					client.t.Errorf("%s", err)
				}

				if err := client.signalSessionDescription(&answer); err != nil {
					client.t.Errorf("%s", err)
				}

				if err := client.conn.SetLocalDescription(answer); err != nil {
					client.t.Errorf("%s", err)
				}

				// Protect the client.pendingCandidates variable
				client.mu.Lock()
				func() {
					defer client.mu.Unlock()
				}()

				// 4. Notify remote server of any pending candidates
				for _, c := range client.pendingCandidates {
					if err = client.signalCandidate(c); err != nil {
						client.t.Logf("%s", err)
					}
				}

			// Event Type 3
			// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
			case *pb.SignalingEvent_RtcIceCandidateInit:

				// 1. Convert from protobuf to pion/webrtc
				rtcIceCand_pb := msg.GetRtcIceCandidateInit()

				// Convert from protobuf to pion/webrtc
				var rtcIceCand_pion webrtc.ICECandidateInit
				if err := zutils.ConvertFromProtoMessage(rtcIceCand_pb.ProtoReflect(), &rtcIceCand_pion); err != nil {
					client.t.Errorf("%s", err)
				}

				// 2. Add a remote ICE candidate
				if err := client.conn.AddICECandidate(rtcIceCand_pion); err != nil {
					client.t.Logf("%s", err)
				}

			case nil:
				client.t.Logf("Signaling event message with empty Event field")
			default:
				client.t.Logf("SignalingEvent.Event has unexpected type %T", evt)
			}
		}
	}
}

// register handlers for the RTCPeerConnection
func (client *testClient) initClientHandlers() {

	// When an ICE candidate is available send to the "server" instance
	client.conn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		client.mu.Lock()
		defer client.mu.Unlock()

		// If the remote description has not been set, add to the list of pending candidates
		desc := client.conn.RemoteDescription()
		if desc == nil {

			client.pendingCandidates = append(client.pendingCandidates, c)

		} else {

			if err := client.signalCandidate(c); err != nil {
				client.t.Logf("%s", err)
			}

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
func (client *testClient) initEchoDataChannel() {
	echo_impl, err := client.conn.CreateDataChannel(Echo.String(), dcConfigs[Echo])
	if err != nil {
		client.t.Errorf("%s", err)
	}

	echo, err := NewDataChannel(Echo, echo_impl)
	zutils.FailOnError(err, "Error creating data channel interface: ")

	client.dataChannels[Echo] = echo // this is used in the Send(DataChannelLabel, proto.Message) function
	client.Receiver <- echo.Updates()
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
