package webrtc

import (
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

// DataChannel with echo
func dataChannelEcho(w http.ResponseWriter, r *http.Request) {

	var (
		c   *websocket.Conn
		ws  zws.WebSocket
		rtc WebRTC
	)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	ws = zws.NewWebSocket(c)
	rtc, err = NewWebRTC(ws)
	if err != nil {
		log.Println(err)
		return
	}
	defer rtc.Close()

	ch := rtc.DataChannel(Echo)
	finish := make(chan bool, 1)
	go func() {
		time.Sleep(10 * time.Second)
		close(finish)
	}()

	for {
		select {
		case _, ok := <-finish:
			if !ok {
				return
			}
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
func TestDataChannel(test *testing.T) {
	// Create test server with the echo handler.
	s := httptest.NewServer(http.HandlerFunc(dataChannelEcho))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	WS := zws.NewWebSocket(ws)
	defer WS.Close()

	// WebRTC Connection
	var conn *webrtc.PeerConnection

	// an example of what a browser client might do
	receiver := WS.Updates()
	mu := &sync.Mutex{}
	localCandidates := make([]*webrtc.ICECandidate, 0)
	var pendingSdp *webrtc.SessionDescription
	for {
		select {
		case msg, ok := <-receiver:
			if !ok {
				return
			}

			evt := msg.(*pb.SignalingEvent).GetEvent()

			switch t := evt.(type) {

			case *pb.SignalingEvent_RtcIceServer:
				test.Logf("stun server")
				iceServer_pb := msg.(*pb.SignalingEvent).GetRtcIceServer()

				var iceServer_pion webrtc.ICEServer
				err = zutils.ConvertFromProtoMessage(iceServer_pb.ProtoReflect(), &iceServer_pion)
				if err != nil {
					test.Errorf("%s", err)
				}

				var config = webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{iceServer_pion},
				}

				conn, err = webrtc.NewPeerConnection(config)
				if err != nil {
					test.Errorf("%s", err)
				}

				// When an ICE candidate is available send to the "server" instance
				conn.OnICECandidate(func(c *webrtc.ICECandidate) {
					if c == nil {
						return
					}

					mu.Lock()
					defer mu.Unlock()

					desc := conn.RemoteDescription()
					if desc == nil {
						localCandidates = append(localCandidates, c)
					} else {
						iceCandInit := c.ToJSON()

						var cand pb.RTCIceCandidateInit
						err := zutils.ConvertToProtoMessage(&iceCandInit, cand.ProtoReflect())
						if err != nil {
							test.Logf("Error converting proto message: %s", err)
							return
						}

						msg := pb.SignalingEvent{
							Event: &pb.SignalingEvent_RtcIceCandidateInit{
								RtcIceCandidateInit: &cand,
							},
						}

						_ = WS.Send(msg.ProtoReflect().Interface())
					}
				})

				// Set the handler for ICE connection state
				// This will notify you when the peer has connected/disconnected
				conn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
					test.Logf("Client ICE Connection State has changed: %s\n", connectionState.String())
				})

				// Register data channel creation handling
				conn.OnDataChannel(func(d *webrtc.DataChannel) {
					test.Logf("New DataChannel %s %d\n", d.Label(), d.ID())

					// Register channel opening handling
					d.OnOpen(func() {
						test.Logf("Data channel '%s'-'%d' open. Random messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())

						for range time.NewTicker(1 * time.Second).C {
							test_msg := pb.Echo{}

							var b []byte
							b, err := proto.Marshal(&test_msg)
							if err != nil {
								return
							}

							// Send the message as text
							test.Logf("Sending %s", test_msg.String())
							err = d.Send(b)
							if err != nil {
								test.Errorf(err.Error())
							}
						}
					})

					// Register text message handling
					d.OnMessage(func(msg webrtc.DataChannelMessage) {
						test.Logf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
					})
				})

				mu.Lock()
				func() {
					defer mu.Unlock()
				}()

				if pendingSdp != nil {
					err = conn.SetRemoteDescription(*pendingSdp)
					if err != nil {
						test.Errorf("%s", err)
					}

					answer, err := conn.CreateAnswer(nil)
					if err != nil {
						test.Errorf("%s", err)
					}

					signalSessionDescription(test, WS, answer)

					err = conn.SetLocalDescription(answer)
					if err != nil {
						test.Errorf("%s", err)
					}

					for _, c := range localCandidates {
						signalCandidate(WS, c)
					}
				}

			case *pb.SignalingEvent_SessionDescription:
				test.Logf("session description")
				sdp_pb := msg.(*pb.SignalingEvent).GetSessionDescription()

				var sdp_pion webrtc.SessionDescription
				_ = zutils.ConvertFromProtoMessage(sdp_pb.ProtoReflect(), &sdp_pion)

				if conn == nil {
					pendingSdp = &sdp_pion
					continue
				}

				err = conn.SetRemoteDescription(sdp_pion)
				if err != nil {
					test.Logf("%s", err)
					continue
				}

				answer, err := conn.CreateAnswer(nil)
				if err != nil {
					test.Errorf("%s", err)
				}

				signalSessionDescription(test, WS, answer)

				err = conn.SetLocalDescription(answer)
				if err != nil {
					test.Errorf("%s", err)
				}

				mu.Lock()
				func() {
					defer mu.Unlock()
				}()
				for _, c := range localCandidates {
					signalCandidate(WS, c)
				}

			case *pb.SignalingEvent_RtcIceCandidateInit:
				test.Logf("ice candidate")
				rtcIceCand_pb := msg.(*pb.SignalingEvent).GetRtcIceCandidateInit()

				var rtcIceCand_pion webrtc.ICECandidateInit
				_ = zutils.ConvertFromProtoMessage(rtcIceCand_pb.ProtoReflect(), &rtcIceCand_pion)

				err = conn.AddICECandidate(rtcIceCand_pion)
				zutils.WarnOnError(err, "Error adding ICE Candidate: ")

			case nil:
				test.Logf("Signaling event message with empty Event field")
			default:
				test.Logf("SignalingEvent.Event has unexpected type %T", t)
			}
		}
	}
}

func signalCandidate(ws zws.WebSocket, c *webrtc.ICECandidate) {
	iceCandInit := c.ToJSON()

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

	_ = ws.Send(msg.ProtoReflect().Interface())
}

func signalSessionDescription(t *testing.T, ws zws.WebSocket, sdp webrtc.SessionDescription) {
	var sdp_pb pb.SessionDescription
	err := zutils.ConvertToProtoMessage(&sdp, sdp_pb.ProtoReflect())
	if err != nil {
		t.Errorf("%s", err)
	}

	msg := pb.SignalingEvent{
		Event: &pb.SignalingEvent_SessionDescription{
			SessionDescription: &sdp_pb,
		},
	}

	err = ws.Send(msg.ProtoReflect().Interface())
	if err != nil {
		t.Errorf("%s", err)
	}
}
