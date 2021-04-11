package game

import (
	"context"
)

type GameType int

const (
	TestGame GameType = iota + 1
)

func (typ GameType) String() string {
	return [...]string{"", "TestGame"}[typ]
}

type mediaStreamType int

const (
	TestVP8 mediaStreamType = iota + 1
	TestOpus
)

func (typ mediaStreamType) String() string {
	return [...]string{"", "TestVP8", "TestOpus"}[typ]
}

type key int

const (
	Port key = iota
)

// this will panic if the type assertion fails
func fromContext(ctx context.Context) (int, bool) {
	port, ok := ctx.Value(Port).(int)
	return port, ok
}
