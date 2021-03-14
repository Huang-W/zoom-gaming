package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	websocket "github.com/gorilla/websocket"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	pb "zoomgaming/proto"
	zoomws "zoomgaming/websocket"
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
}

func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"Game server is alive!"})
	}
}

// WebSocket init
func websocketHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		var (
			wsConn  *websocket.Conn
			webSocket zoomws.WebSocket
		)

		wsConn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("upgrading http request: %s", err)
			return
		}

		webSocket = zoomws.NewWebSocket(wsConn)

		// test1
		wsMsg := pb.SignalingEvent{
			Event: &pb.SignalingEvent_RtcIceServer{
				RtcIceServer: &pb.RTCIceServer{
					Urls: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}
		webSocket.Send(wsMsg.ProtoReflect().Interface())
	}
}
