package websocket

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	websocket "github.com/gorilla/websocket"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {

	var (
		c        *websocket.Conn
		ws       WebSocket
		receiver <-chan []byte
	)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	ws = NewWebSocket(c)
	defer ws.Close()
	receiver = ws.Updates()
	if err != nil {
		return
	}

	for {
		select {
		case msg, ok := <-receiver:
			if !ok {
				return
			}
			_ = ws.Send(msg)
		}
	}
}

func initialize(f http.HandlerFunc) (s *httptest.Server, ws *websocket.Conn, err error) {
	// Create test server with the echo handler.
	s = httptest.NewServer(f)

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	ws, _, err = websocket.DefaultDialer.Dial(u, nil)

	return
}

func TestEcho(t *testing.T) {

	s, ws, err := initialize(echo)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer s.Close()
	defer func() {
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}()

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 5; i++ {
		msg := pb.SignalingEvent{
			Event: &pb.SignalingEvent_RtcIceServer{
				RtcIceServer: &pb.RTCIceServer{
					Urls: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}

		var request []byte
		request, err := proto.Marshal(&msg)
		if err != nil {
			return
		}

		if err := ws.WriteMessage(websocket.BinaryMessage, request); err != nil {
			t.Fatalf("%v", err)
		}

		var b []byte
		_, b, err = ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}

		echo := &pb.SignalingEvent{}
		err = proto.Unmarshal(b, echo)
		if err != nil {
			t.Fatalf("%v", err)
		}

		// check equality
		if !proto.Equal(msg.ProtoReflect().Interface(), echo) {
			t.Fatalf("Want equality between s1 and s2")
		}
	}
}
