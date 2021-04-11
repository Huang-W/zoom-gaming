package webrtc

import (
	webrtc "github.com/pion/webrtc/v3"

	pref "google.golang.org/protobuf/reflect/protoreflect"

	pb "zoomgaming/proto"
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

var defaultRTCConfiguration = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

var dcConfigs = map[DataChannelLabel](*webrtc.DataChannelInit){
	Echo: &webrtc.DataChannelInit{
		Ordered:    func(b bool) *bool { return &b }(true),
		Negotiated: func(b bool) *bool { return &b }(true),
		ID:         func(i uint16) *uint16 { return &i }(0),
	},
}

var mapping = map[DataChannelLabel](pref.MessageType){
	Echo: (*pb.Echo)(nil).ProtoReflect().Type(),
}

var reverseMapping = map[pref.MessageType](DataChannelLabel){
	(*pb.Echo)(nil).ProtoReflect().Type(): Echo,
}
