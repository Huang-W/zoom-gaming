package room

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/pion/webrtc/v3"

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
	ctx         context.Context
	cancel context.CancelFunc
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

	ctx, cancel := context.WithCancel(context.Background())

	vctx := context.WithValue(ctx, game.Port, 5004)
	actx := context.WithValue(ctx, game.Port, 4004)

	var audioStream game.Stream
	var videoStream game.Stream

	switch typ {
	case game.TestGame:
		videoStream, err = game.NewStream(vctx, game.TestH264)
		utils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = game.NewStream(actx, game.TestOpus)
		utils.FailOnError(err, "Error starting audio stream: %s")
	default:
		videoStream, err = game.NewStream(vctx, game.VideoSH)
		utils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = game.NewStream(actx, game.AudioSH)
		utils.FailOnError(err, "Error starting audio stream: %s")
	}

	g, err := game.NewGame(typ, ctx)
	utils.FailOnError(err, "Error creating game: ")

	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "GameStream")
	utils.FailOnError(err, "Error creating video track: ")

	// Create an audio track
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "GameStream")
	utils.FailOnError(err, "Error creating audio track: ")

	r := &room{
		game:        g,
		ctx:         ctx,
		cancel: cancel,
		audioTrack:  audioTrack,
		videoTrack:  videoTrack,
		audioStream: audioStream,
		videoStream: videoStream,
		mu:          &sync.Mutex{},
		players:     make(map[game.PlayerIndex](rtc.WebRTC), len(game.GameMappings[typ])),
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
	if len(r.players) > 0 {
		return errors.New("Unable to switch game, players still present in room")
	} else {

		typ := game.GameTypeOf(game_id)
		if typ == game.GameUndefined {
			return errors.New("Unable to switch game, invalid game_id")
		}

		r.mu.Lock()
		defer r.mu.Unlock()

		r.game.Close()

		g, err := game.NewGame(typ, r.ctx)
		if err != nil {
			return err
		}

		r.game = g

		return nil
	}
}

func (r *room) NewPlayer(ws ws.WebSocket) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	var idx game.PlayerIndex
	players := []game.PlayerIndex{game.Player1, game.Player2, game.Player3, game.Player4}

	for _, p := range players {
		_, prs := r.players[p]
		if !prs {
			idx = p
			break
		}
	}

	if idx == 0 {
		return errors.New("Only 4 players at a time")
	}

	rtc, err := rtc.NewWebRTC(ws, r.videoTrack, r.audioTrack)
	if err != nil {
		return err
	}

	// CHANGE THIS: Use the first data channel (GameInput) as input for game
	dcs := rtc.DataChannels()
	go func() {
		defer r.removePlayer(idx) // remove player if the rtc connection shuts down
		for ch := range dcs {
			r.game.AttachInputStream(ch, idx)
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

	return nil
}

func (r *room) Close() {
	r.cancel()
}

func (r *room) removePlayer(idx game.PlayerIndex) {

	r.mu.Lock()
	defer r.mu.Unlock()

	_, prs := r.players[idx]
	if prs {
		delete(r.players, idx)
	}

	if prs && idx == game.Player1 {
		for _, conn := range r.players {
			conn.Close()
		}
	}
}
