package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	pj "google.golang.org/protobuf/encoding/protojson"
	pref "google.golang.org/protobuf/reflect/protoreflect"

	webrtc "github.com/pion/webrtc/v3"

	pb "zoomgaming/signaling"
)

var (
	// protobuf marshaling options
	mo = pj.MarshalOptions{
		UseEnumNumbers: false,
	}
	umo = pj.UnmarshalOptions{}

	// mapping of protobuf types to pion/webrtc types
	mapping = map[pref.MessageType]reflect.Type{
		(*pb.RTCIceServer)(nil).ProtoReflect().Type(): reflect.TypeOf((*webrtc.ICEServer)(nil)),
		(*pb.SessionDescription)(nil).ProtoReflect().Type(): reflect.TypeOf((*webrtc.SessionDescription)(nil)),
		(*pb.RTCIceCandidateInit)(nil).ProtoReflect().Type(): reflect.TypeOf((*webrtc.ICECandidateInit)(nil)),
	}
)

func ConvertFromProtoMessage(m pref.Message, dest interface{}) error {

	var (
		// used to compare types
		actualType reflect.Type = reflect.TypeOf(dest)
		expectedType reflect.Type

		// protobuf
		proto_message_type pref.MessageType = m.Type()
	)

	expectedType, present := mapping[proto_message_type]
	if !present {

		return errors.New(fmt.Sprintf("Unsupported protobuf type of %T", proto_message_type.Zero()))

	} else if expectedType != actualType {

		return errors.New(fmt.Sprintf("Type mismatch - Expected: %s - Actual: %s", expectedType.String(), actualType.String()))
	}

	// protobuf message in wire form
	b, err := mo.Marshal(m.Interface())
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, dest)
	if err != nil {
		return err
	}

	return nil
}

func ConvertToProtoMessage(orig interface{}, m pref.Message) error {

	var (
		// used to compare types
		actualType reflect.Type = reflect.TypeOf(orig)
		expectedType reflect.Type

		// protobuf
		proto_message_type pref.MessageType = m.Type()
	)

	expectedType, present := mapping[proto_message_type]
	if !present {

		return errors.New(fmt.Sprintf("Unsupported protobuf type of %T", proto_message_type.Zero()))

	} else if expectedType != actualType {

		return errors.New(fmt.Sprintf("Type mismatch - Expected: %s - Actual: %s", expectedType.String(), actualType.String()))
	}

	// protobuf message in wire form
	b, err := json.Marshal(orig)
	if err != nil {
		return err
	}

	err = pj.Unmarshal(b, m.Interface())
	if err != nil {
		return err
	}

	return nil
}
