package websocket

import (
	"errors"
	"fmt"
	"log"
	"time"

	websocket "github.com/gorilla/websocket"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
	utils "zoomgaming/utils"
)

/**

This file can be used to send/receive messages across/from a websocket connection.
The supported binary format is protobuf.

Current implementation does not confirm whether a message has been received by the client.
// https://blog.golang.org/context

*/

type WebSocket interface {
	Send(m proto.Message) error    // send a message to the browser
	Updates() <-chan proto.Message // a stream of messages from the browser
	Close() error                  // shutdown gracefully
}

type webSocket struct {
	conn *websocket.Conn

	// Send any incoming websocket messages to the channels in receivers
	receiver chan proto.Message
	outbound chan proto.Message
	closing  chan chan error
}

// Call this before using
func NewWebSocket(conn *websocket.Conn) WebSocket {
	ws := &webSocket{
		conn:     conn,
		receiver: make(chan proto.Message, 32),
		outbound: make(chan proto.Message, 32),
		closing:  make(chan chan error),
	}
	go ws.readPump()
	go ws.writePump()
	return ws
}

// Send a message across the internal websocket channel
func (ws *webSocket) Send(m proto.Message) error {
	if ws.outbound == nil || len(ws.outbound) == cap(ws.outbound) {
		return errors.New("Unable to send message")
	}
	ws.outbound <- m
	return nil
}

// Create a receive-only channel that cannot be closed by the requester
func (ws *webSocket) Updates() <-chan proto.Message {
	return ws.receiver
}

func (ws *webSocket) Close() error {
	errc := make(chan error)
	ws.closing <- errc
	return <-errc
}

// readPump forwards messages received from the websocket connection to any registered receivers.
//
// There is at most one reader per websocket connection
func (ws *webSocket) readPump() {
	defer func() {
		ws.conn.Close()
		log.Println("closing ws conn")
	}()
	ws.conn.SetReadLimit(4096)
	ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.conn.SetPongHandler(func(string) error {
		ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	var pending []proto.Message
	var err     error
	for {
		var first proto.Message
		var receiver chan proto.Message
		if len(pending) > 0 {
			first = pending[0]
			receiver = ws.receiver // enable send case
		}

		select {
		case errc := <-ws.closing:
			errc <- err
			close(ws.receiver)
			return
		case receiver <- first:
			pending = pending[1:]
		default:
			// blocks internally until message read or error
			_, b, err := ws.conn.ReadMessage()
			if err != nil {
				// Log an error if this websocket connection did not close properly
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					utils.WarnOnError(err, "Closing Error: ")
				}
				break
			}

			msg := &pb.WebSocketMessage{}
			err = proto.Unmarshal(b, msg)
			if err != nil {
				utils.WarnOnError(err, fmt.Sprintf("Error unmarshaling byte array %v", b))
				break
			}

			pending = append(pending, msg)
		}
	}
}

// writePump pushes queued messages across the websocket connection.
//
// This go-routine ensures there is at most one writer for the websocket connection
//
// A ticker is used for the websocket heartbeat
func (ws *webSocket) writePump() {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		ws.conn.Close()
		log.Println("closing ws conn")
	}()
	var err error
	for {
		select {
		case errc := <-ws.closing:
			errc <- err
			// set outbound channel to nil to prevent future sends
			ws.outbound = nil
			return
		case <-ticker.C:
			ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) // keeps the websocket connection alive
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		// marshal any outbound messages into wireform and send across websocket connection
		case message, ok := <-ws.outbound:
			if !ok {
				return
			}

			w, err := ws.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					utils.WarnOnError(err, "Closing Error: ")
				}
				return
			}

			pbMessage, err := proto.Marshal(message)
			if err != nil {
				utils.WarnOnError(err, fmt.Sprintf("Error marshaling message: %v", message))
				continue
			}
			w.Write(pbMessage)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
