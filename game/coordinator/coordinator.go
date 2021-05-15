package coordinator

import (
	"errors"
	"sync"

	"zoomgaming/game"
	"zoomgaming/room"
	ws "zoomgaming/websocket"
)

type RoomCoordinator interface {
	JoinRoom(string, string, ws.WebSocket) error
}

type roomCoordinator struct {
	rooms    map[string](room.Room)
	mu       *sync.Mutex
	occupancy map[int]bool
	maxRooms int
}

func NewRoomCoordinator(maxRooms int) (res RoomCoordinator, err error) {

	occupancy := make(map[int]bool)
	for i := 0; i < maxRooms; i++ {
		occupancy[i] = false
	}

	c := &roomCoordinator{
		rooms:    make(map[string]room.Room),
		mu:       &sync.Mutex{},
		occupancy: occupancy,
		maxRooms: maxRooms,
	}

	res = c
	return
}

func (c *roomCoordinator) JoinRoom(room_id string, game_id string, ws ws.WebSocket) error {

	c.mu.Lock()
	defer c.mu.Unlock()

	typ := game.GameTypeOf(game_id)
	if typ == game.GameUndefined {
		return errors.New("game type")
	}

	var r room.Room
	var prs bool
	var err error

	r, prs = c.rooms[room_id]
	if !prs {
		if len(c.rooms) >= c.maxRooms {
			return errors.New("max rooms")
		}

		var i int
		for i = 0; i < c.maxRooms; i++ {
			filled := c.occupancy[i]
			if !filled {
				break
			}
		}

		c.occupancy[i] = true

		r, err = room.NewRoom(typ, i)
		if err != nil {
			return err
		}

		c.rooms[room_id] = r
		go func(room_id string, i int) {
			select {
			case <-r.Done():
				delete(c.rooms, room_id)
				c.occupancy[i] = false
			}
		}(room_id, i)
	}

	return r.NewPlayer(ws)
}
