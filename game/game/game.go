package game

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	proto "google.golang.org/protobuf/proto"

	zutils "zoomgaming/utils"
)

/**

  A game is associated with 1 video stream
    and 1 audio stream

*/

type Game interface {
	AudioStream() chan (<-chan []byte)
	VideoStream() chan (<-chan []byte)
	AttachInputStream(<-chan proto.Message) error // mux input streams and relay to the game
	Start() (error, error)                        // start the pair of audio / video streams
	Close()
}

type game struct {
	ctx         context.Context
	audioStream Stream
	videoStream Stream

	wg          *sync.WaitGroup // protects merged player input channels
	playerInput chan proto.Message
	cancel      context.CancelFunc
}

// This must be called
func NewGame(typ GameType, pool *sync.Pool) (g Game, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
			return
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	var audioStream Stream
	var videoStream Stream

	switch typ {
	case TestGame:
		vctx := context.WithValue(ctx, Port, 5004)
		actx := context.WithValue(ctx, Port, 4004)
		videoStream, err = NewStream(vctx, TestVP8, pool)
		zutils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = NewStream(actx, TestOpus, pool)
		zutils.FailOnError(err, "Error starting audio stream: %s")
	default:
		panic("Invalid game type")
	}

	game := &game{
		ctx:         ctx,
		audioStream: audioStream,
		videoStream: videoStream,
		wg:          &sync.WaitGroup{},
		playerInput: make(chan proto.Message, 1024),
		cancel:      cancel,
	}

	go game.feedInput()

	g = game
	return
}

func (g *game) AudioStream() chan (<-chan []byte) {
	return g.audioStream.Updates()
}

func (g *game) VideoStream() chan (<-chan []byte) {
	return g.videoStream.Updates()
}

func (g *game) AttachInputStream(ch <-chan proto.Message) error {

	if err := g.ctx.Err(); err != nil {
		log.Println(err)
		return err
	}

	go func() {

		g.wg.Add(1)
		defer g.wg.Done()

		for msg := range ch {
			g.playerInput <- msg
		}
	}()

	return nil
}

func (g *game) Start() (erra error, errb error) {
	erra = nil // g.audioStream.Start()
	errb = g.videoStream.Start()
	return
}

// context cancel for the vidoe, audio streams
func (g *game) Close() {
	g.cancel()
}

func (g *game) feedInput() {

	defer func() {
		g.wg.Wait()
		close(g.playerInput)
	}()

	for {
		select {
		case <-g.ctx.Done():
			return
		case msg, ok := <-g.playerInput:
			if !ok {
				return
			}
			log.Println(msg) // do something later
		}
	}
}
