package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	"zoomgaming/game"
	"zoomgaming/room"
	zws "zoomgaming/websocket"
)

// Upgrade an HTTP connection to WebSocket
var upgrader = websocket.Upgrader{
	// CORS
	CheckOrigin: func(r *http.Request) bool { return true },
}

var r room.Room

// http server
func NewServer() *negroni.Negroni {
	var err error
	r, err = room.NewRoom(game.TestGame, 0)
	if err != nil {
		log.Printf("Failed to create a room :%s", err)
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
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/demo", roomHandler(formatter)).Methods("GET")
	// mx.HandleFunc("/rooms/{room_id:[a-zA-Z0-9]+}/{gane_id:[a-zA-Z0-9]+}", roomHandler(formatter)).Methods("GET")
}

func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"Game server is alive!"})
	}
}

func roomHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("upgrading http request: %s", err)
			return
		}

		ws := zws.NewWebSocket(conn)

		err = r.NewPlayer(ws)
		if err != nil {
			log.Printf("adding new player to r: %s", err)
		}
	}
}
