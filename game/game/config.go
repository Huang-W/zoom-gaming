package game

import (
	"context"

	x "github.com/linuxdeepin/go-x11-client"
	"github.com/linuxdeepin/go-x11-client/util/keysyms"

	pb "zoomgaming/proto"
)

type GameType int

const (
	GameUndefined GameType = iota
	TestGame
	SpaceTime
)

func (typ GameType) String() string {
	return [...]string{"", "TestGame", "SpaceTime"}[typ]
}

func GameTypeOf(game_id string) GameType {
	if SpaceTime.String() == game_id {
		return SpaceTime
	} else {
		return GameUndefined
	}
}

type keysymMapping map[pb.KeyPressEvent_Key](x.Keysym)
type keycodeMapping map[pb.KeyPressEvent_Key](x.Keycode)
type gameMapping map[PlayerIndex](keysymMapping)

var GameMappings = map[GameType](gameMapping){
	TestGame: gameMapping{
		Player1: keysymMapping{},
		Player2: keysymMapping{},
		Player3: keysymMapping{},
		Player4: keysymMapping{},
	},
	SpaceTime: gameMapping{
		Player1: keysymMapping{
			pb.KeyPressEvent_KEY_ARROW_LEFT:  keysyms.XK_Left,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: keysyms.XK_Right,
			pb.KeyPressEvent_KEY_ARROW_UP:    keysyms.XK_Up,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  keysyms.XK_Down,
			pb.KeyPressEvent_KEY_SPACE:       keysyms.XK_space,
			pb.KeyPressEvent_KEY_KEY_D:       keysyms.XK_D,
			pb.KeyPressEvent_KEY_KEY_S:       keysyms.XK_S,
			pb.KeyPressEvent_KEY_KEY_A:       keysyms.XK_A,
		},
		Player2: keysymMapping{
			pb.KeyPressEvent_KEY_ARROW_LEFT:  keysyms.XK_Q,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: keysyms.XK_W,
			pb.KeyPressEvent_KEY_ARROW_UP:    keysyms.XK_E,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  keysyms.XK_R,
			pb.KeyPressEvent_KEY_SPACE:       keysyms.XK_T,
			pb.KeyPressEvent_KEY_KEY_D:       keysyms.XK_Y,
			pb.KeyPressEvent_KEY_KEY_S:       keysyms.XK_U,
			pb.KeyPressEvent_KEY_KEY_A:       keysyms.XK_I,
		},
		Player3: keysymMapping{
			pb.KeyPressEvent_KEY_ARROW_LEFT:  keysyms.XK_1,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: keysyms.XK_2,
			pb.KeyPressEvent_KEY_ARROW_UP:    keysyms.XK_3,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  keysyms.XK_4,
			pb.KeyPressEvent_KEY_SPACE:       keysyms.XK_5,
			pb.KeyPressEvent_KEY_KEY_D:       keysyms.XK_6,
			pb.KeyPressEvent_KEY_KEY_S:       keysyms.XK_7,
			pb.KeyPressEvent_KEY_KEY_A:       keysyms.XK_8,
		},
		Player4: keysymMapping{
			pb.KeyPressEvent_KEY_ARROW_LEFT:  keysyms.XK_Z,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: keysyms.XK_X,
			pb.KeyPressEvent_KEY_ARROW_UP:    keysyms.XK_C,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  keysyms.XK_V,
			pb.KeyPressEvent_KEY_SPACE:       keysyms.XK_B,
			pb.KeyPressEvent_KEY_KEY_D:       keysyms.XK_N,
			pb.KeyPressEvent_KEY_KEY_S:       keysyms.XK_M,
			pb.KeyPressEvent_KEY_KEY_A:       keysyms.XK_comma,
		},
	},
}

type mediaStreamType int

const (
	TestH264 mediaStreamType = iota + 1
	TestOpus
	VideoSH
	AudioSH
)

func (typ mediaStreamType) String() string {
	return [...]string{"", "TestVP8", "TestOpus", "VideoSH", "AudioSH"}[typ]
}

type key int

const (
	Port key = iota
)

type PlayerIndex int

const (
	Player1 PlayerIndex = iota + 1
	Player2
	Player3
	Player4
)

func (idx PlayerIndex) String() string {
	return [...]string{"", "Player1", "Player2", "Player3", "Player4"}[idx]
}

// this will panic if the type assertion fails
func fromContext(ctx context.Context) (int, bool) {
	port, ok := ctx.Value(Port).(int)
	return port, ok
}
