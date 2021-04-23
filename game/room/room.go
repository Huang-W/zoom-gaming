package room

import (
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
	NewPlayer(ws.WebSocket) error
}

type room struct {
	game       game.Game
	audioTrack *webrtc.TrackLocalStaticRTP
	videoTrack *webrtc.TrackLocalStaticRTP
	// playerTracks []*webrtc.TrackLocalStaticRTP

	mu      *sync.Mutex
	players map[game.PlayerIndex](rtc.WebRTC)
}

func NewRoom() (res Room, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	g, err := game.NewGame(game.TestGame)
	utils.FailOnError(err, "Error creating game: ")

	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "GameStream")
	utils.FailOnError(err, "Error creating video track: ")

	// Create an audio track
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "GameStream")
	utils.FailOnError(err, "Error creating audio track: ")

	r := &room{
		game:       g,
		audioTrack: audioTrack,
		videoTrack: videoTrack,
		mu:         &sync.Mutex{},
		players:    make(map[game.PlayerIndex](rtc.WebRTC)),
	}

	go func() {
		select {
		case ch := <-r.game.AudioStream():
			go func() {
				for pckt := range ch {
					r.audioTrack.Write(pckt)
				}
			}()
		}
	}()

	go func() {
		select {
		case ch := <-r.game.VideoStream():
			go func() {
				for pckt := range ch {
					r.videoTrack.Write(pckt)
				}
			}()
		}
	}()

	erra, errb := r.game.Start()
	if erra != nil || errb != nil {
		err = errors.New(fmt.Sprintf("Unable to start game -- video: %s -- audio: %s", erra, errb))
	}

	res = r
	return
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

func (r *room) removePlayer(idx game.PlayerIndex) {

	r.mu.Lock()
	defer r.mu.Unlock()

	_, prs := r.players[idx]
	if prs {
		delete(r.players, idx)
	}
}
