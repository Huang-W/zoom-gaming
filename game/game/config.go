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
	return [...]string{"", "TestGame"}[typ]
}

type mediaStreamType int

const (
	TestVP8 mediaStreamType = iota + 1
	TestOpus
	X11VP8
	PulseOpus
	VideoSH
	AudioSH
)

func (typ mediaStreamType) String() string {
	return [...]string{"", "TestVP8", "TestOpus", "X11VP8", "PulseOpus", "VideoSH", "AudioSH"}[typ]
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
