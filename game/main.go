package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	"zoomgaming/coordinator"
	zws "zoomgaming/websocket"
)

var addr = flag.String("addr", ":8080", "http service address")
var c coordinator.RoomCoordinator

func main() {

	flag.Parse()
	var err error

	c, err = coordinator.NewRoomCoordinator(2)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	server := NewServer()
	server.Run(*addr)
}

// Upgrade an HTTP connection to WebSocket
var upgrader = websocket.Upgrader{
	// CORS
	CheckOrigin: func(r *http.Request) bool { return true },
}

// http server
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
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	s := mx.PathPrefix("/demo").Subrouter()
	s.HandleFunc("", gameHandler(formatter)).Methods("GET")
	s.HandleFunc("/{room_id}/{game_id}", gameHandler(formatter)).Methods("GET")
	// mx.HandleFunc("/rooms/{room_id:[a-zA-Z0-9]+}/{gane_id:[a-zA-Z0-9]+}", roomHandler(formatter)).Methods("GET")
}

func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"Game server is alive!"})
	}
}

func gameHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		vars := mux.Vars(req)

		room_id, prs := vars["room_id"]
		if !prs {
			room_id = "1111"
		}

		game_id, prs := vars["game_id"]
		if !prs {
			game_id = "SpaceTime"
		}

		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("upgrading http request: %s", err)
			return
		}

		ws := zws.NewWebSocket(conn)

		if err := c.JoinRoom(room_id, game_id, ws); err != nil {
			log.Printf("joining room: %s", err)
			ws.Close()
		}
	}
}
