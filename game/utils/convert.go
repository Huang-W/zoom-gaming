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
	mapping = map[reflect.Type]reflect.Type{
		reflect.TypeOf((*pb.RTCIceServer)(nil)): reflect.TypeOf((*webrtc.ICEServer)(nil)),
		reflect.TypeOf((*pb.SessionDescription)(nil)): reflect.TypeOf((*webrtc.SessionDescription)(nil)),
		reflect.TypeOf((*pb.RTCIceCandidateInit)(nil)): reflect.TypeOf((*webrtc.ICECandidateInit)(nil)),
	}
)

func ConvertFromProtoMessage(m pref.Message, dest interface{}) error {

	var (
		// used to compare types
		actualType reflect.Type = reflect.TypeOf(dest)
		expectedType reflect.Type

		// protobuf
		proto_message pref.ProtoMessage = m.Interface()
		proto_message_type reflect.Type = reflect.TypeOf(proto_message)
	)

	expectedType, present := mapping[proto_message_type]
	if !present {

		return errors.New(fmt.Sprintf("Unsupport protobuf type of %T", proto_message_type))

	} else if expectedType != actualType {

		return errors.New(fmt.Sprintf("Type mismatch - Expected: %T - Actual: %T", expectedType, actualType))
	}

	// protobuf message in wire form
	b, err := mo.Marshal(proto_message)
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
		proto_message pref.ProtoMessage = m.Interface()
		proto_message_type reflect.Type = reflect.TypeOf(proto_message)
	)

	expectedType, present := mapping[proto_message_type]
	if !present {

		return errors.New(fmt.Sprintf("Unsupport protobuf type of %T", proto_message_type))

	} else if expectedType != actualType {

		return errors.New(fmt.Sprintf("Type mismatch - Expected: %T - Actual: %T", expectedType, actualType))
	}

	// protobuf message in wire form
	b, err := json.Marshal(orig)
	if err != nil {
		return err
	}

	err = pj.Unmarshal(b, proto_message)
	if err != nil {
		return err
	}

	return nil
}
