package game

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"

	uinput "gopkg.in/bendahl/uinput.v1"

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
	typ      GameType
	gameExec *exec.Cmd
	cancel   context.CancelFunc
}

func NewGame(typ GameType) (g Game, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
			return
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	var gameExec *exec.Cmd

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

	gameExec.Env = append(os.Environ(), fmt.Sprintf("DISPLAY=:%d", 99))

	game := &game{
		typ:      typ,
		gameExec: gameExec,
		cancel:   cancel,
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
		}
	}()

	g = game
	return
}

func (g *game) AttachInputStream(ch <-chan proto.Message, idx PlayerIndex, keyboard uinput.Keyboard) error {

	mapping, prs := GameMappings[g.typ][idx]
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
			for msg := range ch {
				switch t := msg.(type) {
				case *pb.InputEvent:
					evt := msg.(*pb.InputEvent).GetKeyPressEvent()
					key := mapping[evt.GetKey()]
					switch evt.GetDirection() {
					case pb.KeyPressEvent_DIRECTION_UP:
						keyboard.KeyUp(key)
					case pb.KeyPressEvent_DIRECTION_DOWN:
						keyboard.KeyDown(key)
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

// context cancel for the vidoe, audio streams
func (g *game) Stop() {
	g.cancel()
}
