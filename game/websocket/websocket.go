package websocket

import (
	"errors"
	"fmt"
	"io/ioutil"
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

*/

type WebSocket struct {
	conn *websocket.Conn

	// Send any incoming websocket messages to the channels in receivers
	receivers [](chan proto.Message)
	outbound  chan proto.Message
}

// Call this before using
func NewWebSocket(conn *websocket.Conn) *WebSocket {
	ws := &WebSocket{
		conn: conn,
		receivers: make([](chan proto.Message), 0),
		outbound: make(chan proto.Message, 256),
	}
	go ws.readPump()
	go ws.writePump()
	return ws
}

// Send a message across the internal websocket channel
func (ws *WebSocket) Send(m proto.Message) error {
	if ws.outbound == nil {
		// ws has not been initialized
		return errors.New("Initialize a new WebSocket using the constructor")
	}
	ws.outbound <- m
	return nil
}

// Create a receive-only channel that cannot be closed by the requester
func (ws *WebSocket) CreateReceiver() (<-chan proto.Message, error) {
	if ws.receivers == nil {
		// ws has not been initialized
		return nil, errors.New("Initialize a new WebSocket using the constructor")
	}
	receiver := make(chan proto.Message, 256)
	ws.receivers = append(ws.receivers, receiver)
	return receiver, nil
}

func (ws *WebSocket) Stop() {
	if ws.conn == nil {
		return
	}
	ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

// readPump forwards messages received from the websocket connection to any registered receivers.
//
// There is at most one reader per websocket connection
func (ws *WebSocket) readPump() {
	defer func() {
		ws.conn.Close()
		// close all channels that might still be expecting a message
		for _, ch := range ws.receivers {
			close(ch)
		}
		log.Println("closing ws conn")
	}()
	ws.conn.SetReadLimit(4096)
	ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.conn.SetPongHandler(func(string) error {
		ws.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {

		_, r, err := ws.conn.NextReader()
    if err != nil {
			// Log an error if this websocket connection did not close properly
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				utils.WarnOnError(err, "Closing Error: ")
			}
      break
    }

		b, err := ioutil.ReadAll(r)
		if err != nil {
			utils.WarnOnError(err, fmt.Sprintf("Error reading byte array %v", b))
			continue
		}

		msg := &pb.WebSocketMessage{}
		err = proto.Unmarshal(b, msg)
		if err != nil {

			utils.WarnOnError(err, fmt.Sprintf("Error unmarshaling byte array %v", b))

		} else {

			for _, ch := range ws.receivers {
				ch <- msg.ProtoReflect().Interface()
			}
		}
	}
}

// writePump pushes queued messages across the websocket connection.
//
// This go-routine ensures there is at most one writer for the websocket connection
//
// A ticker is used for the websocket heartbeat
func (ws *WebSocket) writePump() {
	ticker := time.NewTicker(50 * time.Second)
	defer func() {
		ticker.Stop()
		ws.conn.Close()
		// set outbound channel to nil to prevent future sends
		close(ws.outbound)
		ws.outbound = nil
		log.Println("closing ws conn")
	}()
	for {
		select {
		// keeps the websocket connection alive
		case <-ticker.C:
			ws.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ws.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		// marshal any outbound messages into wireform and send across websocket connection
		case message, ok := <-ws.outbound:
			if !ok {
				// if not ok, the channel has been closed
				ws.Stop()
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
				// any errors with the writer are permanent
				return
			}
		}
	}
}
