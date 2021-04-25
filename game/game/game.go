package game

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"sync"

	x "github.com/linuxdeepin/go-x11-client"
	"github.com/linuxdeepin/go-x11-client/ext/test"
	"github.com/linuxdeepin/go-x11-client/util/keysyms"

	proto "google.golang.org/protobuf/proto"

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
	ctx         context.Context
	typ         GameType
	gameExec    *exec.Cmd
	audioStream Stream
	videoStream Stream
	xdisplay *x.Conn

	wg     *sync.WaitGroup // protects merged player input channels
	cancel context.CancelFunc

	mu             *sync.Mutex
	occupancy      map[PlayerIndex](bool)
	player1Mapping map[pb.KeyPressEvent_Key](x.Keycode)
	player2Mapping map[pb.KeyPressEvent_Key](x.Keycode)
	player3Mapping map[pb.KeyPressEvent_Key](x.Keycode)
	player4Mapping map[pb.KeyPressEvent_Key](x.Keycode)
}

// This must be called
func NewGame(typ GameType, roomIndex int) (g Game, err error) {

	var xdisplay *x.Conn

	defer func() {
		if r := recover(); r != nil {
			if xdisplay != nil {
				xdisplay.Close()
			}
			err = errors.New(fmt.Sprintf("%s", r))
			return
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	var gameExec *exec.Cmd
	var audioStream Stream
	var videoStream Stream

	vctx := context.WithValue(ctx, Port, 5004)
	actx := context.WithValue(ctx, Port, 4004)

	curr_user, err := user.Current()
	if err != nil {
		panic("Error looking up current user")
	}

	switch typ {
	case TestGame:
		gameExec = &exec.Cmd{Path: ""}
		videoStream, err = NewStream(vctx, TestH264, roomIndex)
		zutils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = NewStream(actx, TestOpus, roomIndex)
		zutils.FailOnError(err, "Error starting audio stream: %s")
	case SpaceTime:
		gameExec = exec.CommandContext(ctx, fmt.Sprintf("%s/games/SpaceTime/start.sh", curr_user.HomeDir))
		videoStream, err = NewStream(vctx, VideoSH, roomIndex)
		zutils.FailOnError(err, "Error starting video stream: %s")
		audioStream, err = NewStream(actx, AudioSH, roomIndex)
		zutils.FailOnError(err, "Error starting audio stream: %s")
	default:
		panic("Invalid game type")
	}

	gameExec.Env = append(os.Environ(), fmt.Sprintf("DISPLAY=:%d", 99-roomIndex))
	xdisplay, err = x.NewConnDisplay(fmt.Sprintf(":%d", 99-roomIndex))
	if err != nil {
		// try connecting to :0 instead
		xdisplay, err = x.NewConnDisplay(":0")
		if err != nil {
			panic(fmt.Sprintf("Unable to connect to display %d and :0", 99-roomIndex))
		}
	}

	var occupancy = map[PlayerIndex](bool){
		Player1: false,
		Player2: false,
		Player3: false,
		Player4: false,
	}

	symbols := keysyms.NewKeySymbols(xdisplay)

	var p1mapping = map[pb.KeyPressEvent_Key](x.Keycode){
		pb.KeyPressEvent_KEY_ARROW_LEFT:  symbols.GetKeycodes(keysyms.XK_Left)[0],
		pb.KeyPressEvent_KEY_ARROW_RIGHT: symbols.GetKeycodes(keysyms.XK_Right)[0],
		pb.KeyPressEvent_KEY_ARROW_UP:    symbols.GetKeycodes(keysyms.XK_Up)[0],
		pb.KeyPressEvent_KEY_ARROW_DOWN:  symbols.GetKeycodes(keysyms.XK_Down)[0],
		pb.KeyPressEvent_KEY_SPACE:       symbols.GetKeycodes(keysyms.XK_space)[0],
		pb.KeyPressEvent_KEY_KEY_D:       symbols.GetKeycodes(keysyms.XK_D)[0],
		pb.KeyPressEvent_KEY_KEY_S:       symbols.GetKeycodes(keysyms.XK_S)[0],
		pb.KeyPressEvent_KEY_KEY_A:       symbols.GetKeycodes(keysyms.XK_A)[0],
	}

	var p2mapping = map[pb.KeyPressEvent_Key](x.Keycode){
		pb.KeyPressEvent_KEY_ARROW_LEFT:  symbols.GetKeycodes(keysyms.XK_Q)[0],
		pb.KeyPressEvent_KEY_ARROW_RIGHT: symbols.GetKeycodes(keysyms.XK_W)[0],
		pb.KeyPressEvent_KEY_ARROW_UP:    symbols.GetKeycodes(keysyms.XK_E)[0],
		pb.KeyPressEvent_KEY_ARROW_DOWN:  symbols.GetKeycodes(keysyms.XK_R)[0],
		pb.KeyPressEvent_KEY_SPACE:       symbols.GetKeycodes(keysyms.XK_T)[0],
		pb.KeyPressEvent_KEY_KEY_D:       symbols.GetKeycodes(keysyms.XK_Y)[0],
		pb.KeyPressEvent_KEY_KEY_S:       symbols.GetKeycodes(keysyms.XK_U)[0],
		pb.KeyPressEvent_KEY_KEY_A:       symbols.GetKeycodes(keysyms.XK_I)[0],
	}

	var p3mapping = map[pb.KeyPressEvent_Key](x.Keycode){
		pb.KeyPressEvent_KEY_ARROW_LEFT:  symbols.GetKeycodes(keysyms.XK_1)[0],
		pb.KeyPressEvent_KEY_ARROW_RIGHT: symbols.GetKeycodes(keysyms.XK_2)[0],
		pb.KeyPressEvent_KEY_ARROW_UP:    symbols.GetKeycodes(keysyms.XK_3)[0],
		pb.KeyPressEvent_KEY_ARROW_DOWN:  symbols.GetKeycodes(keysyms.XK_4)[0],
		pb.KeyPressEvent_KEY_SPACE:       symbols.GetKeycodes(keysyms.XK_5)[0],
		pb.KeyPressEvent_KEY_KEY_D:       symbols.GetKeycodes(keysyms.XK_6)[0],
		pb.KeyPressEvent_KEY_KEY_S:       symbols.GetKeycodes(keysyms.XK_7)[0],
		pb.KeyPressEvent_KEY_KEY_A:       symbols.GetKeycodes(keysyms.XK_8)[0],
	}

	var p4mapping = map[pb.KeyPressEvent_Key](x.Keycode){
		pb.KeyPressEvent_KEY_ARROW_LEFT:  symbols.GetKeycodes(keysyms.XK_Z)[0],
		pb.KeyPressEvent_KEY_ARROW_RIGHT: symbols.GetKeycodes(keysyms.XK_X)[0],
		pb.KeyPressEvent_KEY_ARROW_UP:    symbols.GetKeycodes(keysyms.XK_C)[0],
		pb.KeyPressEvent_KEY_ARROW_DOWN:  symbols.GetKeycodes(keysyms.XK_V)[0],
		pb.KeyPressEvent_KEY_SPACE:       symbols.GetKeycodes(keysyms.XK_B)[0],
		pb.KeyPressEvent_KEY_KEY_D:       symbols.GetKeycodes(keysyms.XK_N)[0],
		pb.KeyPressEvent_KEY_KEY_S:       symbols.GetKeycodes(keysyms.XK_M)[0],
		pb.KeyPressEvent_KEY_KEY_A:       symbols.GetKeycodes(keysyms.XK_comma)[0],
	}

	game := &game{
		typ:         typ,
		ctx:         ctx,
		gameExec:    gameExec,
		audioStream: audioStream,
		videoStream: videoStream,
		xdisplay:       xdisplay,
		wg:             &sync.WaitGroup{},
		cancel:         cancel,
		mu:             &sync.Mutex{},
		occupancy:      occupancy,
		player1Mapping: p1mapping,
		player2Mapping: p2mapping,
		player3Mapping: p3mapping,
		player4Mapping: p4mapping,
	}

	if game.gameExec.Path != "" {
		err = game.gameExec.Start()
		if err != nil {
			panic("Error starting game")
		}
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

	var mapping map[pb.KeyPressEvent_Key](x.Keycode)
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

			root := g.xdisplay.GetDefaultScreen().Root

			for msg := range ch {
				// log.Println(msg)
				switch t := msg.(type) {
				case *pb.InputEvent:
					evt := msg.(*pb.InputEvent).GetKeyPressEvent()
					key := mapping[evt.GetKey()]
					switch evt.GetDirection() {
					case pb.KeyPressEvent_DIRECTION_UP:
						test.FakeInput(g.xdisplay, x.KeyReleaseEventCode, uint8(key), x.CurrentTime, root, 0, 0, 0)
					case pb.KeyPressEvent_DIRECTION_DOWN:
						test.FakeInput(g.xdisplay, x.KeyPressEventCode, uint8(key), x.CurrentTime, root, 0, 0, 0)
					default:
						log.Printf("No direction specified")
						continue
					}
					g.xdisplay.Flush()
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
		// g.virtualKeyboard.Close()
	}()

	select {
	case <-g.ctx.Done():
		return
	}
}
