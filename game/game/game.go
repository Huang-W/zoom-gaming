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
)

/**

  A game is associated with 1 video stream
    and 1 audio stream

*/

type Game interface {
	AttachInputStream(<-chan proto.Message, PlayerIndex) error // mux input streams and relay to the game
	Close()
}

type game struct {
	rctx     context.Context // room context
	gctx     context.Context // game context (for switching games)
	typ      GameType
	gameExec *exec.Cmd
	xdisplay *x.Conn

	wg     *sync.WaitGroup // protects merged player input channels
	cancel context.CancelFunc

	mu             *sync.Mutex
	occupancy      map[PlayerIndex](bool)
	playerMappings map[PlayerIndex](keycodeMapping)
}

// This must be called
func NewGame(typ GameType, rctx context.Context) (g Game, err error) {

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

	gctx, cancel := context.WithCancel(rctx)

	var gameExec *exec.Cmd

	curr_user, err := user.Current()
	if err != nil {
		panic("Error looking up current user")
	}

	switch typ {
	case TestGame:
		gameExec = &exec.Cmd{Path: ""}
	case SpaceTime:
		gameExec = exec.CommandContext(gctx, fmt.Sprintf("%s/games/SpaceTime/start.sh", curr_user.HomeDir))
	default:
		panic("Invalid game type")
	}

	gameExec.Env = append(os.Environ(), fmt.Sprintf("DISPLAY=:%d", 99))
	xdisplay, err = x.NewConnDisplay(fmt.Sprintf(":%d", 99))
	if err != nil {
		// try connecting to :0 instead
		xdisplay, err = x.NewConnDisplay(":0")
		if err != nil {
			panic(fmt.Sprintf("Unable to connect to display %d and :0", 99))
		}
	}

	symbols := keysyms.NewKeySymbols(xdisplay)

	keysymMappings := GameMappings[typ]
	keycodeMappings := make(map[PlayerIndex](keycodeMapping), len(keysymMappings))
	for player, mapping := range keysymMappings {
		keycodeMappings[player] = make(keycodeMapping, len(mapping))
		for keyPressEvent, keysym := range mapping {
			keycodeMappings[player][keyPressEvent] = symbols.GetKeycodes(keysym)[0]
		}
	}

	game := &game{
		typ:            typ,
		rctx:           rctx,
		gctx:           gctx,
		gameExec:       gameExec,
		xdisplay:       xdisplay,
		wg:             &sync.WaitGroup{},
		cancel:         cancel,
		mu:             &sync.Mutex{},
		occupancy:      make(map[PlayerIndex](bool)),
		playerMappings: keycodeMappings,
	}

	for key := range game.playerMappings {
		game.occupancy[key] = false
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

func (g *game) AttachInputStream(ch <-chan proto.Message, idx PlayerIndex) error {

	if err := g.gctx.Err(); err != nil {
		log.Println(err)
		return err
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	_, prs := g.occupancy[idx]
	if !prs {
		return errors.New("This player's seat is occupied")
	}

	mapping, prs := g.playerMappings[idx]
	if !prs {
		return errors.New("player not found")
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

// context cancel for the vidoe, audio streams
func (g *game) Close() {
	g.cancel()
}

func (g *game) awaitCancel() {

	defer func() {
		g.wg.Wait()
		g.Close()
		// g.virtualKeyboard.Close()
	}()

	select {
	case <-g.gctx.Done():
		return
	case <-g.rctx.Done():
		return
	}
}
