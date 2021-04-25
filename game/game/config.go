package game

import (
	"context"
)

type GameType int

const (
	TestGame GameType = iota + 1
	SpaceTime
)

func (typ GameType) String() string {
	return [...]string{"", "TestGame", "SpaceTime", "PacPong"}[typ]
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
