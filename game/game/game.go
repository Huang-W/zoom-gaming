package game

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"

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
	Stop()
}

type game struct {
	typ            GameType
	xdisplay       *x.Conn
	gameExec       *exec.Cmd
	cancel         context.CancelFunc
	playerMappings map[PlayerIndex](keycodeMapping)
}

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
	xdisplay, err = x.NewConnDisplay(fmt.Sprintf(":%d", 99-roomIndex))
	if err != nil {
		panic(fmt.Sprintf("Unable to connect to display %d", 99-roomIndex))
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

	curr_user, err := user.Current()
	if err != nil {
		panic("Error looking up current user")
	}

	switch typ {
	case TestGame:
		gameExec = &exec.Cmd{Path: ""}
	case SpaceTime:
		gameExec = exec.CommandContext(ctx, fmt.Sprintf("%s/games/SpaceTime/game/LoversInADangerousSpacetime.x86_64", curr_user.HomeDir))
		gameExec.Dir = fmt.Sprintf("%s/games/SpaceTime/game/", curr_user.HomeDir)
	case Broforce:
		gameExec = exec.CommandContext(ctx, fmt.Sprintf("%s/games/Broforce/game/Broforce.x86_64", curr_user.HomeDir))
		gameExec.Dir = fmt.Sprintf("%s/games/Broforce/game/", curr_user.HomeDir)
	default:
		panic("Invalid game type")
	}

	gameExec.Env = append(os.Environ(), fmt.Sprintf("DISPLAY=:%d", 99-roomIndex))

	game := &game{
		typ:            typ,
		xdisplay:       xdisplay,
		gameExec:       gameExec,
		cancel:         cancel,
		playerMappings: keycodeMappings,
	}

	if game.gameExec.Path != "" {
		err = game.gameExec.Start()
		if err != nil {
			panic("Error starting game")
		}
	}

	go func() {
		gameExec.Wait()
	}()

	go func() {
		select {
		case <-ctx.Done():
			gameExec.Process.Signal(os.Interrupt)
			game.xdisplay.Close()
		}
	}()

	g = game
	return
}

func (g *game) AttachInputStream(ch <-chan proto.Message, idx PlayerIndex) error {

	mapping, prs := g.playerMappings[idx]
	if !prs {
		return errors.New("player not found")
	}

	if g.typ == TestGame {
		go func() {
			for msg := range ch {
				msg := msg.(*pb.InputEvent)
				log.Printf("Received msg from player %s: %s", idx, msg)
			}
		}()
	} else {
		go func() {
			root := g.xdisplay.GetDefaultScreen().Root
			for msg := range ch {
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
func (g *game) Stop() {
	g.cancel()
}
