package utils

import (
	"encoding/json"
	"errors"
	"fmt"

	pj "google.golang.org/protobuf/encoding/protojson"
	pref "google.golang.org/protobuf/reflect/protoreflect"

	webrtc "github.com/pion/webrtc/v3"

	pb "zoomgaming/signaling"
)

var (
	mo = pj.MarshalOptions{
		UseEnumNumbers: false,
	}
	umo = pj.UnmarshalOptions{}
)

func ConvertFromProtoMessage(m pref.Message, dest interface{}) error {

	var (
		b   []byte
		err error
		// plumbing
		proto_message pref.ProtoMessage = m.Interface()
	)

	switch t := proto_message.(type) {
	case *pb.RTCIceServer:
		break
	case *pb.SessionDescription:
		break
	case *pb.RTCIceCandidateInit:
		break
	default:
		err = errors.New(fmt.Sprintf("Unsupported type %T", t))
		return err
	}

	if b, err = mo.Marshal(proto_message); err != nil {
		return err
	}

	if err = json.Unmarshal(b, dest); err != nil {
		return err
	}

	return nil
}

func ConvertToProtoMessage(orig interface{}, m pref.Message) error {
	var (
		b   []byte
		err error
		// plumbing
		proto_message pref.ProtoMessage = m.Interface()
	)

	switch t := orig.(type) {
	case *webrtc.ICEServer:
		break
	case *webrtc.SessionDescription:
		break
	case *webrtc.ICECandidateInit:
		break
	default:
		err = errors.New(fmt.Sprintf("Unsupported type %T", t))
		return err
	}

	if b, err = json.Marshal(orig); err != nil {
		return err
	}

	if err = pj.Unmarshal(b, proto_message); err != nil {
		return err
	}

	return nil
}
