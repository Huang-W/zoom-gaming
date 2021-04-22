package game

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	proto "google.golang.org/protobuf/proto"

	uinput "gopkg.in/bendahl/uinput.v1"

	pb "zoomgaming/proto"
	zutils "zoomgaming/utils"
)

/**

  A game is associated with 1 video stream
    and 1 audio stream

*/

type Game interface {
	AudioStream() chan (<-chan []byte)
	VideoStream() chan (<-chan []byte)
	AttachInputStream(<-chan proto.Message, PlayerIndex) error // mux input streams and relay to the game
	Start() (error, error)                                     // start the pair of audio / video streams
	Close()
}

type game struct {
	ctx             context.Context
	typ GameType
	audioStream     Stream
	videoStream     Stream
	virtualKeyboard uinput.Keyboard

	wg     *sync.WaitGroup // protects merged player input channels
	cancel context.CancelFunc

	mu             *sync.Mutex
	occupancy      map[PlayerIndex](bool)
	player1Mapping map[pb.KeyPressEvent_Key](int)
	player2Mapping map[pb.KeyPressEvent_Key](int)
	player3Mapping map[pb.KeyPressEvent_Key](int)
	player4Mapping map[pb.KeyPressEvent_Key](int)
}

// This must be called
func NewGame(typ GameType) (g Game, err error) {

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
		videoStream, err = NewStream(vctx, TestH264)
		zutils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = NewStream(actx, TestOpus)
		zutils.FailOnError(err, "Error starting audio stream: %s")
	case SpaceTime:
		vctx := context.WithValue(ctx, Port, 5004)
		actx := context.WithValue(ctx, Port, 4004)
		videoStream, err = NewStream(vctx, VideoSH)
		zutils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = NewStream(actx, AudioSH)
		zutils.FailOnError(err, "Error starting audio stream: %s")
	default:
		panic("Invalid game type")
	}

	keyboard, err := uinput.CreateKeyboard("/dev/uinput", []byte("virtualkeyboard"))
	if err != nil {
		panic("Unable to create virtual keyboard")
	}

	var occupancy = map[PlayerIndex](bool){
		Player1: false,
		Player2: false,
		Player3: false,
		Player4: false,
	}

	var p1mapping = map[pb.KeyPressEvent_Key](int){
		pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyLeft,
		pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyRight,
		pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyUp,
		pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyDown,
		pb.KeyPressEvent_KEY_SPACE:       uinput.KeySpace,
		pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyD,
		pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyS,
		pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyA,
	}

	var p2mapping = map[pb.KeyPressEvent_Key](int){
		pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyQ,
		pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyW,
		pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyE,
		pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyR,
		pb.KeyPressEvent_KEY_SPACE:       uinput.KeyT,
		pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyY,
		pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyU,
		pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyI,
	}

	var p3mapping = map[pb.KeyPressEvent_Key](int){
		pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.Key1,
		pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.Key2,
		pb.KeyPressEvent_KEY_ARROW_UP:    uinput.Key3,
		pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.Key4,
		pb.KeyPressEvent_KEY_SPACE:       uinput.Key5,
		pb.KeyPressEvent_KEY_KEY_D:       uinput.Key6,
		pb.KeyPressEvent_KEY_KEY_S:       uinput.Key7,
		pb.KeyPressEvent_KEY_KEY_A:       uinput.Key8,
	}

	var p4mapping = map[pb.KeyPressEvent_Key](int){
		pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyZ,
		pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyX,
		pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyC,
		pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyV,
		pb.KeyPressEvent_KEY_SPACE:       uinput.KeyB,
		pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyN,
		pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyM,
		pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyComma,
	}

	game := &game{
		typ: typ,
		ctx:             ctx,
		audioStream:     audioStream,
		videoStream:     videoStream,
		virtualKeyboard: keyboard,
		wg:              &sync.WaitGroup{},
		cancel:          cancel,
		mu:              &sync.Mutex{},
		occupancy:       occupancy,
		player1Mapping:  p1mapping,
		player2Mapping:  p2mapping,
		player3Mapping:  p3mapping,
		player4Mapping:  p4mapping,
	}

	go game.awaitCancel()

	g = game
	return
}

func (g *game) AudioStream() chan (<-chan []byte) {
	return g.audioStream.Updates()
}

func (g *game) VideoStream() chan (<-chan []byte) {
	return g.videoStream.Updates()
}

func (g *game) AttachInputStream(ch <-chan proto.Message, idx PlayerIndex) error {

	if err := g.ctx.Err(); err != nil {
		log.Println(err)
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	_, prs := g.occupancy[idx]
	if !prs {
		return errors.New("This player's seat is occupied")
	}

	var mapping map[pb.KeyPressEvent_Key](int)
	switch idx {
	case Player1:
		mapping = g.player1Mapping
	case Player2:
		mapping = g.player2Mapping
	case Player3:
		mapping = g.player3Mapping
	case Player4:
		mapping = g.player4Mapping
	default:
		return errors.New("player  not found")
	}

	if g.typ == TestGame {
		go func() {

			g.wg.Add(1)
			defer g.wg.Done()

			for msg := range ch {
				msg := msg.(*pb.InputEvent)
				log.Printf("Received msg from player %s: %s", idx, msg)
			}
		}()
	} else {
		go func() {

			g.wg.Add(1)
			defer g.wg.Done()

			for msg := range ch {
				// log.Println(msg)
				switch t := msg.(type) {
				case *pb.InputEvent:
					evt := msg.(*pb.InputEvent).GetKeyPressEvent()
					key := mapping[evt.GetKey()]
					switch evt.GetDirection() {
					case pb.KeyPressEvent_DIRECTION_UP:
						g.virtualKeyboard.KeyUp(key)
					case pb.KeyPressEvent_DIRECTION_DOWN:
						g.virtualKeyboard.KeyDown(key)
					default:
						log.Printf("No direction specified")
						continue
					}
				default:
					log.Printf("Unexcepted type: %T", t)
					continue
				}
			}
		}()
	}

	return nil
}

func (g *game) Start() (erra error, errb error) {
	erra = g.audioStream.Start()
	errb = g.videoStream.Start()
	return
}

// context cancel for the vidoe, audio streams
func (g *game) Close() {
	g.cancel()
}

func (g *game) awaitCancel() {

	defer func() {
		g.wg.Wait()
		g.virtualKeyboard.Close()
	}()

	select {
	case <-g.ctx.Done():
		return
	}
}
