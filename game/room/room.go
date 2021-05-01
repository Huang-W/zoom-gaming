package room

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/pion/webrtc/v3"

	uinput "gopkg.in/bendahl/uinput.v1"

	game "zoomgaming/game"
	utils "zoomgaming/utils"
	rtc "zoomgaming/webrtc"
	ws "zoomgaming/websocket"
)

/**

A room is associated with 1 game at a time
Up to 4 players in a room

// At some point, the players should have some identifying information beyond just an RTCPeerConnection

*/

type Room interface {
	SwitchGame(string) error
	NewPlayer(ws.WebSocket) error
	Close()
}

type room struct {
	game        game.Game
	typ         game.GameType
	audioTrack  *webrtc.TrackLocalStaticRTP // the game's audio track, shared between all players
	videoTrack  *webrtc.TrackLocalStaticRTP // the game's video track, shared between all players
	audioStream game.Stream
	videoStream game.Stream
	// playerTracks []*webrtc.TrackLocalStaticRTP

	mu      *sync.Mutex // protects players
	players map[game.PlayerIndex](rtc.WebRTC)
}

func NewRoom(typ game.GameType) (res Room, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	var audioStream game.Stream
	var videoStream game.Stream

	switch typ {
	case game.TestGame:
		videoStream, err = game.NewStream(game.TestH264)
		utils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = game.NewStream(game.TestOpus)
		utils.FailOnError(err, "Error starting audio stream: %s")
	default:
		videoStream, err = game.NewStream(game.VideoSH)
		utils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = game.NewStream(game.AudioSH)
		utils.FailOnError(err, "Error starting audio stream: %s")
	}

	g, err := game.NewGame(typ)
	utils.FailOnError(err, "Error creating game: ")

	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "GameStream")
	utils.FailOnError(err, "Error creating video track: ")

	// Create an audio track
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "GameStream")
	utils.FailOnError(err, "Error creating audio track: ")

	r := &room{
		game:        g,
		typ:         typ,
		audioTrack:  audioTrack,
		videoTrack:  videoTrack,
		audioStream: audioStream,
		videoStream: videoStream,
		mu:          &sync.Mutex{},
		players:     make(map[game.PlayerIndex](rtc.WebRTC)),
	}

	go func() {
		select {
		case ch := <-r.audioStream.Updates():
			go func() {
				for pckt := range ch {
					r.audioTrack.Write(pckt)
				}
			}()
		}
	}()

	go func() {
		select {
		case ch := <-r.videoStream.Updates():
			go func() {
				for pckt := range ch {
					r.videoTrack.Write(pckt)
				}
			}()
		}
	}()

	res = r
	return
}

func (r *room) SwitchGame(game_id string) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	typ := game.GameTypeOf(game_id)

	if typ == game.GameUndefined || r.typ == typ {
		return errors.New("Unable to switch game...")
	} else {

		r.game.Stop()

		g, err := game.NewGame(typ)
		if err != nil {
			return err
		}

		for _, conn := range r.players {
			conn.Close()
		}
		r.players = make(map[game.PlayerIndex](rtc.WebRTC))

		r.game = g
		r.typ = typ

		return nil
	}
}

func (r *room) NewPlayer(ws ws.WebSocket) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	var idx game.PlayerIndex
	mappings := []game.PlayerIndex{game.Player1, game.Player2, game.Player3, game.Player4}
	for _, player := range mappings {
		_, prs := r.players[player]
		if !prs {
			idx = player
			break
		}
	}

	if idx == game.PlayerUndefined {
		return errors.New(fmt.Sprintf("Only %d allowed players at a time", len(r.keycodeMappings)))
	}

	rtc, err := rtc.NewWebRTC(ws, r.videoTrack, r.audioTrack)
	if err != nil {
		return err
	}

	// CHANGE THIS: Use the first data channel (GameInput) as input for game
	dcs := rtc.DataChannels()
	go func() {
		// initialize keyboard and check for possible errors
		keyboard, err := uinput.CreateKeyboard("/dev/uinput", []byte(fmt.Sprintf("%s_virtualkeyboard_%d", r.typ, idx)))
		if err != nil {
			return err
		}
		defer keyboard.Close()
		defer r.removePlayer(idx) // remove player if the rtc connection shuts down
		for ch := range dcs {
			r.game.AttachInputStream(ch, idx, keyboard)
		}
	}()
	/**
	tracks := rtc.Broadcast()
	go func() {
		for track := range tracks {
			r.mu.Lock()
			r.playerTracks.append(track)
			r.mu.Unlock()
		}
	}()
	*/
	r.players[idx] = rtc

	log.Println("number of players in the room after adding: ", len(r.players))

	return nil
}

func (r *room) Close() {
	r.videoStream.Stop()
	r.audioStream.Stop()
}

func (r *room) removePlayer(idx game.PlayerIndex) {

	r.mu.Lock()
	defer r.mu.Unlock()

	_, prs := r.players[idx]
	if prs {
		delete(r.players, idx)
	}
}
