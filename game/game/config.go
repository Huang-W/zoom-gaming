package game

import (
	"github.com/bendahl/uinput"
	_ "github.com/linuxdeepin/go-x11-client"
	_ "github.com/linuxdeepin/go-x11-client/util/keysyms"

	pb "zoomgaming/proto"
)

type GameType int

const (
	GameUndefined GameType = iota
	TestGame
	SpaceTime
	Broforce
)

func (typ GameType) String() string {
	return [...]string{"", "TestGame", "SpaceTime", "Broforce"}[typ]
}

func GameTypeOf(game_id string) GameType {
	if SpaceTime.String() == game_id {
		return SpaceTime
	} else if Broforce.String() == game_id {
		return Broforce
	} else {
		return GameUndefined
	}
}

/**
type keysymMapping map[pb.KeyPressEvent_Key](x.Keysym)
type KeycodeMapping map[pb.KeyPressEvent_Key](x.Keycode)
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
	Broforce: gameMapping{
		Player1: keysymMapping{
			pb.KeyPressEvent_KEY_ARROW_UP:    keysyms.XK_Up,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  keysyms.XK_Down,
			pb.KeyPressEvent_KEY_ARROW_LEFT:  keysyms.XK_Left,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: keysyms.XK_Right,
			pb.KeyPressEvent_KEY_KEY_D:       keysyms.XK_D,
			pb.KeyPressEvent_KEY_KEY_S:       keysyms.XK_S,
			pb.KeyPressEvent_KEY_KEY_A:       keysyms.XK_A,
			pb.KeyPressEvent_KEY_SPACE:       keysyms.XK_space,
		},
		Player2: keysymMapping{
			pb.KeyPressEvent_KEY_ARROW_UP:    keysyms.XK_Q,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  keysyms.XK_W,
			pb.KeyPressEvent_KEY_ARROW_LEFT:  keysyms.XK_E,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: keysyms.XK_R,
			pb.KeyPressEvent_KEY_KEY_D:       keysyms.XK_T,
			pb.KeyPressEvent_KEY_KEY_S:       keysyms.XK_Y,
			pb.KeyPressEvent_KEY_KEY_A:       keysyms.XK_U,
			pb.KeyPressEvent_KEY_SPACE:       keysyms.XK_I,
		},
		Player3: keysymMapping{
			pb.KeyPressEvent_KEY_ARROW_UP:    keysyms.XK_1,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  keysyms.XK_2,
			pb.KeyPressEvent_KEY_ARROW_LEFT:  keysyms.XK_3,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: keysyms.XK_4,
			pb.KeyPressEvent_KEY_KEY_D:       keysyms.XK_5,
			pb.KeyPressEvent_KEY_KEY_S:       keysyms.XK_6,
			pb.KeyPressEvent_KEY_KEY_A:       keysyms.XK_7,
			pb.KeyPressEvent_KEY_SPACE:       keysyms.XK_8,
		},
		Player4: keysymMapping{
			pb.KeyPressEvent_KEY_ARROW_UP:    keysyms.XK_Z,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  keysyms.XK_X,
			pb.KeyPressEvent_KEY_ARROW_LEFT:  keysyms.XK_C,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: keysyms.XK_V,
			pb.KeyPressEvent_KEY_KEY_D:       keysyms.XK_B,
			pb.KeyPressEvent_KEY_KEY_S:       keysyms.XK_N,
			pb.KeyPressEvent_KEY_KEY_A:       keysyms.XK_M,
			pb.KeyPressEvent_KEY_SPACE:       keysyms.XK_comma,
		},
	},
}
*/

type Keycode int
type KeyMapping map[pb.KeyPressEvent_Key](Keycode)
type gameMapping map[PlayerIndex](KeyMapping)

var GameMappings = map[GameType](gameMapping){
	TestGame: gameMapping{
		Player1: KeyMapping{},
		Player2: KeyMapping{},
		Player3: KeyMapping{},
		Player4: KeyMapping{},
	},
	SpaceTime: gameMapping{
		Player1: KeyMapping{
			pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyLeft,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyRight,
			pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyUp,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyDown,
			pb.KeyPressEvent_KEY_SPACE:       uinput.KeySpace,
			pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyD,
			pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyS,
			pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyA,
		},
		Player2: KeyMapping{
			pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyQ,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyW,
			pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyE,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyR,
			pb.KeyPressEvent_KEY_SPACE:       uinput.KeyT,
			pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyY,
			pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyU,
			pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyI,
		},
		Player3: KeyMapping{
			pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.Key1,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.Key2,
			pb.KeyPressEvent_KEY_ARROW_UP:    uinput.Key3,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.Key4,
			pb.KeyPressEvent_KEY_SPACE:       uinput.Key5,
			pb.KeyPressEvent_KEY_KEY_D:       uinput.Key6,
			pb.KeyPressEvent_KEY_KEY_S:       uinput.Key7,
			pb.KeyPressEvent_KEY_KEY_A:       uinput.Key8,
		},
		Player4: KeyMapping{
			pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyZ,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyX,
			pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyC,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyV,
			pb.KeyPressEvent_KEY_SPACE:       uinput.KeyB,
			pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyN,
			pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyM,
			pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyComma,
		},
	},
	Broforce: gameMapping{
		Player1: KeyMapping{
			pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyUp,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyDown,
			pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyLeft,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyRight,
			pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyD,
			pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyS,
			pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyA,
			pb.KeyPressEvent_KEY_SPACE:       uinput.KeySpace,
		},
		Player2: KeyMapping{
			pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyQ,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyW,
			pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyE,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyR,
			pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyT,
			pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyY,
			pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyU,
			pb.KeyPressEvent_KEY_SPACE:       uinput.KeyI,
		},
		Player3: KeyMapping{
			pb.KeyPressEvent_KEY_ARROW_UP:    uinput.Key1,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.Key2,
			pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.Key3,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.Key4,
			pb.KeyPressEvent_KEY_KEY_D:       uinput.Key5,
			pb.KeyPressEvent_KEY_KEY_S:       uinput.Key6,
			pb.KeyPressEvent_KEY_KEY_A:       uinput.Key7,
			pb.KeyPressEvent_KEY_SPACE:       uinput.Key8,
		},
		Player4: KeyMapping{
			pb.KeyPressEvent_KEY_ARROW_UP:    uinput.KeyZ,
			pb.KeyPressEvent_KEY_ARROW_DOWN:  uinput.KeyX,
			pb.KeyPressEvent_KEY_ARROW_LEFT:  uinput.KeyC,
			pb.KeyPressEvent_KEY_ARROW_RIGHT: uinput.KeyV,
			pb.KeyPressEvent_KEY_KEY_D:       uinput.KeyB,
			pb.KeyPressEvent_KEY_KEY_S:       uinput.KeyN,
			pb.KeyPressEvent_KEY_KEY_A:       uinput.KeyM,
			pb.KeyPressEvent_KEY_SPACE:       uinput.KeyComma,
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

type PlayerIndex int

const (
	PlayerUndefined PlayerIndex = iota
	Player1
	Player2
	Player3
	Player4
	Player5
	Player6
	Player7
	Player8
)

func (idx PlayerIndex) String() string {
	return [...]string{"", "Player1", "Player2", "Player3", "Player4", "Player5", "Player6", "Player7", "Player8"}[idx]
}
