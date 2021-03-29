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

	protojson "google.golang.org/protobuf/encoding/protojson"
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
	ch := rtc.DataChannel(Echo)

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
	testClient, err := NewTestClient(t, WS)
	if err != nil {
		t.Errorf("%s", err)
	}

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

	// The "Receive" go routine
	//
	// a separate go routine to log any echoed messages from the server
	go func() {
		receiving := testClient.DataChannel(Echo)
		for {
			select {
			case msg, ok := <-receiving:
				if !ok {
					return
				}
				b, _ := protojson.Marshal(msg)
				t.Logf("Messaged received from server: %s", b)
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
 * The Webrtc connection is established by signaling messages across a websocket connection.
 */
type testClient struct {
	conn *webrtc.PeerConnection
	t    *testing.T

	ws           zws.WebSocket
	receivers    map[DataChannelLabel](chan proto.Message)  // data channel messages arrive here
	dataChannels map[DataChannelLabel](*webrtc.DataChannel) // internal handling of data channels
	flowControls map[DataChannelLabel](chan struct{})       // prevent overflowing of data channel send buffer
	mu           *sync.Mutex                                // protects the pendingCandidates variable

	// After the remote session description is set, the candidates in pendingCandidates will be signaled to the remote agent
	pendingCandidates []*webrtc.ICECandidate

	/**
	 * pendingSdp holds the remote session description in case the
	 * "pb.SessionDescription" message is received before the local RTCPeerConnection is initialized
	 */
	pendingSdp *webrtc.SessionDescription
}

func NewTestClient(t *testing.T, ws zws.WebSocket) (WebRTC WebRTC, err error) {

	// Catch any panics and return (nil, err) when panicing
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	c := &testClient{
		t:                 t,
		ws:                ws,
		receivers:         make(map[DataChannelLabel](chan proto.Message)),
		dataChannels:      make(map[DataChannelLabel](*webrtc.DataChannel)),
		flowControls:      make(map[DataChannelLabel](chan struct{})),
		mu:                &sync.Mutex{},
		pendingCandidates: make([]*webrtc.ICECandidate, 0),
	}

	// create receiver in advance
	c.receivers[Echo] = make(chan proto.Message, 1024)
	c.flowControls[Echo] = make(chan struct{})

	go c.signalingEvents()

	WebRTC = c
	return
}

// Incoming messages for a datachannel
func (client *testClient) DataChannel(label DataChannelLabel) <-chan proto.Message {
	return client.receivers[label]
}

// Send a message across the datachannel
func (client *testClient) Send(label DataChannelLabel, msg proto.Message) error {

	// Marshal the protobuf message into a byte array
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	// Check if the data channel has been created
	dc, prs := client.dataChannels[label]
	if !prs {
		return errors.New("Label not found")
	}

	// Block if the buffer will exceed 1 MB
	if dc.BufferedAmount()+uint64(len(b)) > maxBufferedAmount {
		<-client.flowControls[label]
	}

	// Send the message using the SCTP transport
	err = dc.Send(b)

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
		case msg, ok := <-clientUpdates:
			if !ok {
				// When the websocket channel is closed, start clean-up
				for _, ch := range client.receivers {
					close(ch)
				}
				for _, ch := range client.flowControls {
					close(ch)
				}
				return
			}

			// Make sure the received message is pb.SignalingEvent
			//
			// Else, log to console
			if msg_pb, ok := msg.(*pb.SignalingEvent); ok {

				evt := msg_pb.GetEvent()

				switch t := evt.(type) {

				// Event Type 1
				// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
				case *pb.SignalingEvent_RtcIceServer:

					// 1. Convert from protobuf to pion/webrtc
					iceServer_pb := msg.(*pb.SignalingEvent).GetRtcIceServer()

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
					sdp_pb := msg.(*pb.SignalingEvent).GetSessionDescription()

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
					rtcIceCand_pb := msg.(*pb.SignalingEvent).GetRtcIceCandidateInit()

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
					client.t.Logf("SignalingEvent.Event has unexpected type %T", t)
				}

			} else {
				client.t.Logf("%s", msg)
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
			// Else, send the candidate to the remote agent

			var iceCandInit webrtc.ICECandidateInit
			iceCandInit = c.ToJSON()

			// Wrap the ice candidate as a pb.SignalingEvent
			msg, err := zutils.WrapRTCIceCandidateInit(&iceCandInit)
			if err != nil {
				client.t.Logf("Error converting proto message: %s", err)
				return
			}

			if err := client.ws.Send(msg.ProtoReflect().Interface()); err != nil {
				client.t.Logf("Error sending websocket message: %s", err)
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
	var dc *webrtc.DataChannel
	dc, err := client.conn.CreateDataChannel(Echo.String(), &webrtc.DataChannelInit{
		Ordered:    &echo_ordered,    // in-sequence messages
		Negotiated: &echo_negotiated, // Data channel is either pre-negotiated or we need to fire a "new channel" event
		ID:         &echo_id,         // custom ID specified for pre-negotiated channels only
	})
	if err != nil {
		client.t.Errorf("%s", err)
	}

	client.dataChannels[Echo] = dc // this is used in the Send(DataChannelLabel, proto.Message) function

	// Register text message handling
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {

		b := msg.Data // the message as a byte array

		client.t.Logf("Message from DataChannel '%s': '%s'\n", dc.Label(), b)

		// Unmarshal from protobuf wireform
		var pb_msg pb.Echo
		err := proto.Unmarshal(b, pb_msg.ProtoReflect().Interface())
		if err != nil {
			return
		}

		// Notify any listeners of a new message
		client.receivers[Echo] <- pb_msg.ProtoReflect().Interface()
	})

	// The bufferedAmountLowThreshold notifies us of when we can send more data
	dc.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)

	// This callback is made when the current bufferedAmount becomes lower than the threadshold
	dc.OnBufferedAmountLow(func() {
		client.flowControls[Echo] <- struct{}{}
	})
}

// Helper function to send an RtcIceCandidateInit to the server
// over the websocket connection
func (client *testClient) signalCandidate(c *webrtc.ICECandidate) error {

	// Convert to dictionary form
	var iceCandInit webrtc.ICECandidateInit
	iceCandInit = c.ToJSON()

	// Wrap the ice candidate as a pb.SignalingEvent
	msg, err := zutils.WrapRTCIceCandidateInit(&iceCandInit)
	if err != nil {
		return err
	}

	// Send the message over websocket
	if err := client.ws.Send(msg.ProtoReflect().Interface()); err != nil {
		return err
	}

	return nil
}

// Helper function to send a SessionDescription to the server
// over the websocket connection
func (client *testClient) signalSessionDescription(sdp *webrtc.SessionDescription) error {

	// Wrap the sdp as a pb.SignalingEvent
	msg, err := zutils.WrapSessionDescription(sdp)
	if err != nil {
		return err
	}

	// Send the message over websocket
	if err := client.ws.Send(msg.ProtoReflect().Interface()); err != nil {
		return err
	}

	return nil
}
