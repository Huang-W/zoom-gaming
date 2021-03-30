package utils

import (
	"errors"
	"fmt"

	webrtc "github.com/pion/webrtc/v3"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
)

func MarshalSignalingEvent(i interface{}) ([]byte, error) {

	switch i.(type) {

	case *webrtc.ICEServer:

		iceServer := pb.RTCIceServer{}

		if err := ConvertToProtoMessage(i, iceServer.ProtoReflect()); err != nil {
			return nil, err
		}

		msg := pb.SignalingEvent{
			Event: &pb.SignalingEvent_RtcIceServer{
				RtcIceServer: &iceServer,
			},
		}

		return proto.Marshal(msg.ProtoReflect().Interface())

	case *webrtc.SessionDescription:

		sdp := pb.SessionDescription{}

		if err := ConvertToProtoMessage(i, sdp.ProtoReflect()); err != nil {
			return nil, err
		}

		msg := pb.SignalingEvent{
			Event: &pb.SignalingEvent_SessionDescription{
				SessionDescription: &sdp,
			},
		}

		return proto.Marshal(msg.ProtoReflect().Interface())

	case *webrtc.ICECandidateInit:

		cand := pb.RTCIceCandidateInit{}

		if err := ConvertToProtoMessage(i, cand.ProtoReflect()); err != nil {
			return nil, err
		}

		msg := pb.SignalingEvent{
			Event: &pb.SignalingEvent_RtcIceCandidateInit{
				RtcIceCandidateInit: &cand,
			},
		}

		return proto.Marshal(msg.ProtoReflect().Interface())

	case nil:
		return nil, errors.New("unsupported nil type")
	default:
		return nil, errors.New(fmt.Sprintf("unsupported type %T", i))
	}
}
