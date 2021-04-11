package room

import (
	"errors"
	"fmt"
	"sync"
	"time"

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
	pool *sync.Pool

	mu      *sync.Mutex
	players []rtc.WebRTC
}

func NewRoom() (res Room, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
		}
	}()

	pool := &sync.Pool{
		New: func() interface{} {
			return make([]byte, 1500) // UDP MTU
		},
	}

	g, err := game.NewGame(game.TestGame, pool)
	utils.FailOnError(err, "Error creating game: ")

	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "game-video")
	utils.FailOnError(err, "Error creating video track: ")

	// Create an audio track
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "game-audio")
	utils.FailOnError(err, "Error creating audio track: ")

	r := &room{
		game:       g,
		audioTrack: audioTrack,
		videoTrack: videoTrack,
		pool: pool,
		mu:         &sync.Mutex{},
		players:    make([]rtc.WebRTC, 0),
	}

	res = r
	return
}

func (r *room) NewPlayer(ws ws.WebSocket) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.players) > 3 {
		return errors.New("Only 4 players at a time")
	}

	rtc, err := rtc.NewWebRTC(ws, r.videoTrack, r.audioTrack)
	if err != nil {
		return err
	}

	r.players = append(r.players, rtc)

	// CHANGE THIS: Use the first data channel (Echo) as input for game
	dcs := rtc.DataChannels()
	go func() {
		defer r.removePlayer(rtc)
		for ch := range dcs {
			r.game.AttachInputStream(ch)
		}
	}()

	select {
	case <-rtc.Streaming():
		// first player - start streaming
		if len(r.players) == 1 {

			go func() {
				select {
				case ch := <-r.game.AudioStream():
					go func() {
						for pckt := range ch {
							r.audioTrack.Write(pckt)
							r.pool.Put(pckt)
						}
					}()
				}
			}()

			go func() {
				select {
				case ch := <-r.game.VideoStream():
					go func() {
						defer r.game.Close()
						ticker := time.NewTicker(5 * time.Second)
						defer ticker.Stop()
						var counter int
						var last_count int
						for {
							select {
							case <-ticker.C:
								fmt.Println(counter-last_count)
								last_count = counter
							case pckt, ok := <-ch:
								if !ok {
									return
								}
								r.videoTrack.Write(pckt)
								counter += 1
								r.pool.Put(pckt)
							}
						}
						/**
						for pckt := range ch {
							r.videoTrack.Write(pckt)
							r.pool.Put(pckt)
						}
						*/
					}()
				}
			}()

			erra, errb := r.game.Start()
			if erra != nil || errb != nil {
				rtc.Close()
				return errors.New(fmt.Sprintf("Unable to start game -- video: %s -- audio: %s", erra, errb))
			}
		}
	}

	return nil
}

func (r *room) removePlayer(conn rtc.WebRTC) {

	r.mu.Lock()
	defer r.mu.Unlock()

	players := filter(r.players, func(player rtc.WebRTC) bool { return conn != player }) // remove this player
	r.players = players
}

func filter(conns []rtc.WebRTC, fn func(rtc.WebRTC) bool) []rtc.WebRTC {
	var peers []rtc.WebRTC // == nil
	for _, v := range conns {
		if fn(v) {
			peers = append(peers, v)
		}
	}
	return peers
}
