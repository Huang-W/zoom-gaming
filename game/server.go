package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	websocket "github.com/gorilla/websocket"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
	zwebrtc "zoomgaming/webrtc"
	zws "zoomgaming/websocket"
)

// Upgrade an HTTP connection to WebSocket
//
// https://pkg.go.dev/github.com/gorilla/websocket#Upgrader
var upgrader = websocket.Upgrader{
	// CORS
	CheckOrigin: func(r *http.Request) bool { return true },
}

// http server
//
// https://pkg.go.dev/github.com/urfave/negroni#Negroni
func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})
	n := negroni.Classic()
	mx := mux.NewRouter()
	initRoutes(mx, formatter)
	n.UseHandler(mx)
	return n
}

// REST API routes
//
// https://pkg.go.dev/github.com/gorilla/mux#Router
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/ws", websocketHandler(formatter)).Methods("GET")
	mx.HandleFunc("/webrtc", webrtcHandler(formatter)).Methods("GET")
}

func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"Game server is alive!"})
	}
}

// WebSocket Echo server
func websocketHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var (
			wsConn    *websocket.Conn
			webSocket zws.WebSocket
		)

		wsConn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("upgrading http request: %s", err)
			return
		}

		webSocket = zws.NewWebSocket(wsConn)

		// test - log all received messages to console
		go func() {
			var receiver <-chan []byte
			receiver = webSocket.Updates()
			for {
				select {
				case msg, ok := <-receiver:
					if !ok {
						return
					}
					log.Println("received", msg)
					// echo back to browser
					webSocket.Send(msg)
				}
			}
		}()

		// test1
		wsMsg := &pb.SignalingEvent{
			Event: &pb.SignalingEvent_RtcIceServer{
				RtcIceServer: &pb.RTCIceServer{
					Urls: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}
		b, _ := proto.Marshal(wsMsg)
		webSocket.Send(b)
	}
}

// WebRTC server
// Sends an offer to the browser client
//
// Echoes back any messages reveived on the data chnnale with label of "GameInput"
func webrtcHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var (
			wsConn    *websocket.Conn
			rtcConn   zwebrtc.WebRTC
			webSocket zws.WebSocket
		)

		wsConn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("upgrading http request: %s", err)
			return
		}

		webSocket = zws.NewWebSocket(wsConn)

		rtcConn, err = zwebrtc.NewWebRTC(webSocket)
		if err != nil {
			log.Println(err)
			return
		}

		go func() {
			ch, _ := rtcConn.DataChannel(zwebrtc.Echo)
			for {
				select {
				case msg, ok := <-ch:
					if !ok {
						return
					}
					log.Println(msg)
					// echo back to browser
					rtcConn.Send(zwebrtc.Echo, msg)
				}
			}
		}()
	}
}
