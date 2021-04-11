package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	websocket "github.com/gorilla/websocket"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
	room "zoomgaming/room"
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

var r room.Room

// http server
//
// https://pkg.go.dev/github.com/urfave/negroni#Negroni
func NewServer() *negroni.Negroni {
	var err error
	r, err = room.NewRoom()
	if err != nil {
		os.Exit(1)
	}
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
	mx.HandleFunc("/demo", demoHandler(formatter)).Methods("GET")
}

func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"Game server is alive!"})
	}
}

// WebSocket Echo server
func websocketHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("upgrading http request: %s", err)
			return
		}

		ws := zws.NewWebSocket(conn)

		// test - log all received messages to console
		go func() {
			updates := ws.Updates()
			for ch := range updates {
				for msg := range ch {
					log.Println("received", msg)
					// echo back to browser
					ws.Send(msg)
				}
			}
		}()

		// test1
		msg := &pb.SignalingEvent{
			Event: &pb.SignalingEvent_RtcIceServer{
				RtcIceServer: &pb.RTCIceServer{
					Urls: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}
		b, _ := proto.Marshal(msg)
		ws.Send(b)
	}
}

// WebRTC server
// Sends an offer to the browser client
//
// Echoes back any messages reveived on the data chnnale with label of "Echo"
func webrtcHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("upgrading http request: %s", err)
			return
		}

		ws := zws.NewWebSocket(conn)

		rtc, err := zwebrtc.NewWebRTC(ws, nil, nil)
		if err != nil {
			log.Println(err)
			return
		}

		go func() {
			updates := rtc.DataChannels()
			for ch := range updates {
				for msg := range ch {
					log.Println(msg)
					// echo back to browser
					rtc.Send(msg)
				}
			}
		}()
	}
}

func demoHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("upgrading http request: %s", err)
			return
		}

		ws := zws.NewWebSocket(conn)

		err = r.NewPlayer(ws)
		if err != nil {
			log.Printf("adding new player: %s", err)
		}
	}
}
