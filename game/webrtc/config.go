package webrtc

import (
	webrtc "github.com/pion/webrtc/v3"
)

const (
	bufferedAmountLowThreshold uint64 = 512 * 1024  // 512 KB
	maxBufferedAmount          uint64 = 1024 * 1024 // 1 MB
)

type DataChannelLabel int // Represents a unique data chhanel

const (
	Echo DataChannelLabel = iota + 1
	// GameInput
	// ChatRoom
)

func (label DataChannelLabel) String() string {
	return [...]string{"", "Echo", "GameInput", "ChatRoom"}[label]
}

var defaultConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

// DataChannelInit parameters - Echo
var echo_ordered bool = true
var echo_negotiated bool = true
var echo_id uint16 = 1111
