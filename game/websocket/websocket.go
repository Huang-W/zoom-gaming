package websocket

import (
	"log"
	"sync"
	"time"

	websocket "github.com/gorilla/websocket"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
)

/**

This file can be used to send/receive messages across/from a websocket connection.
The supported binary format is protobuf.

*/

type WebSocket interface {
	Send(proto.Message) error      // send a message to the browser
	Updates() <-chan proto.Message // a stream of messages from the browser
	Close() error                  // try to send a websocket close message
}

type webSocket struct {
	conn *websocket.Conn

	mu       *sync.Mutex        // protect the websocket writer
	receiver chan proto.Message // client messages arrive here
}

// Constructor
func NewWebSocket(conn *websocket.Conn) WebSocket {
	ws := &webSocket{
		conn:     conn,
		mu:       &sync.Mutex{},
		receiver: make(chan proto.Message, 32),
	}

	ws.conn.SetReadLimit(4096)
	ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.conn.SetPongHandler(func(string) error {
		ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	go ws.readPump()
	go ws.heartbeat()
	return ws
}

// Send a message across the internal websocket channel
//
// Only one writer allowed at a time
func (ws *webSocket) Send(m proto.Message) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	w, err := ws.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
			log.Printf("Closing Error: %s", err)
		}
		return err
	}

	pbMessage, err := proto.Marshal(m)
	if err != nil {
		log.Printf("Error marshaling message: %s", err)
		return err
	}

	w.Write(pbMessage)

	if err = w.Close(); err != nil {
		log.Printf("Error closing message: %s", err)
		return err
	}
	return nil
}

// Receive-only channel that cannot be closed by the requester
func (ws *webSocket) Updates() <-chan proto.Message {
	return ws.receiver
}

func (ws *webSocket) Close() (err error) {
	err = ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	return
}

// readPump forwards messages received from the websocket connection
//
// There is at most one reader per websocket connection
func (ws *webSocket) readPump() {
	defer func() {
		ws.conn.Close()
		close(ws.receiver)
		log.Println("closing ws conn")
	}()

	for {
		_, b, err := ws.conn.ReadMessage() // blocks until message read or error
		if err != nil {
			// Log an error if this websocket connection did not close properly
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("Closing Error: %s", err)
			}
			break
		}

		msg := &pb.SignalingEvent{}
		err = proto.Unmarshal(b, msg)
		if err != nil {
			log.Printf("Error unmarshaling byte array %v\nError: %s", b, err)
			continue
		}

		ws.receiver <- msg
	}
}

// Keeps the websocket connection alive
func (ws *webSocket) heartbeat() {

	ticker := time.NewTicker(50 * time.Second)

	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
