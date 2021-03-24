// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: proto/signaling.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// https://pkg.go.dev/github.com/pion/webrtc/v3#SDPType
type SessionDescription_SDPType int32

const (
	SessionDescription_SDP_TYPE_UNSPECIFIED SessionDescription_SDPType = 0
	SessionDescription_offer                SessionDescription_SDPType = 1
	SessionDescription_SDP_TYPE_OFFER       SessionDescription_SDPType = 1
	SessionDescription_pranswer             SessionDescription_SDPType = 2
	SessionDescription_SDP_TYPE_PRANSWER    SessionDescription_SDPType = 2
	SessionDescription_answer               SessionDescription_SDPType = 3
	SessionDescription_SDP_TYPE_ANSWER      SessionDescription_SDPType = 3
	SessionDescription_rollback             SessionDescription_SDPType = 4
	SessionDescription_SDP_TYPE_ROLLBACK    SessionDescription_SDPType = 4
)

// Enum value maps for SessionDescription_SDPType.
var (
	SessionDescription_SDPType_name = map[int32]string{
		0: "SDP_TYPE_UNSPECIFIED",
		1: "offer",
		// Duplicate value: 1: "SDP_TYPE_OFFER",
		2: "pranswer",
		// Duplicate value: 2: "SDP_TYPE_PRANSWER",
		3: "answer",
		// Duplicate value: 3: "SDP_TYPE_ANSWER",
		4: "rollback",
		// Duplicate value: 4: "SDP_TYPE_ROLLBACK",
	}
	SessionDescription_SDPType_value = map[string]int32{
		"SDP_TYPE_UNSPECIFIED": 0,
		"offer":                1,
		"SDP_TYPE_OFFER":       1,
		"pranswer":             2,
		"SDP_TYPE_PRANSWER":    2,
		"answer":               3,
		"SDP_TYPE_ANSWER":      3,
		"rollback":             4,
		"SDP_TYPE_ROLLBACK":    4,
	}
)

func (x SessionDescription_SDPType) Enum() *SessionDescription_SDPType {
	p := new(SessionDescription_SDPType)
	*p = x
	return p
}

func (x SessionDescription_SDPType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SessionDescription_SDPType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_signaling_proto_enumTypes[0].Descriptor()
}

func (SessionDescription_SDPType) Type() protoreflect.EnumType {
	return &file_proto_signaling_proto_enumTypes[0]
}

func (x SessionDescription_SDPType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SessionDescription_SDPType.Descriptor instead.
func (SessionDescription_SDPType) EnumDescriptor() ([]byte, []int) {
	return file_proto_signaling_proto_rawDescGZIP(), []int{2, 0}
}

// A websocket message that represents a certain event
//
// Each event has its own handler
type SignalingEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Event:
	//	*SignalingEvent_RtcIceServer
	//	*SignalingEvent_SessionDescription
	//	*SignalingEvent_RtcIceCandidateInit
	Event isSignalingEvent_Event `protobuf_oneof:"event"`
}

func (x *SignalingEvent) Reset() {
	*x = SignalingEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_signaling_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignalingEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignalingEvent) ProtoMessage() {}

func (x *SignalingEvent) ProtoReflect() protoreflect.Message {
	mi := &file_proto_signaling_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignalingEvent.ProtoReflect.Descriptor instead.
func (*SignalingEvent) Descriptor() ([]byte, []int) {
	return file_proto_signaling_proto_rawDescGZIP(), []int{0}
}

func (m *SignalingEvent) GetEvent() isSignalingEvent_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (x *SignalingEvent) GetRtcIceServer() *RTCIceServer {
	if x, ok := x.GetEvent().(*SignalingEvent_RtcIceServer); ok {
		return x.RtcIceServer
	}
	return nil
}

func (x *SignalingEvent) GetSessionDescription() *SessionDescription {
	if x, ok := x.GetEvent().(*SignalingEvent_SessionDescription); ok {
		return x.SessionDescription
	}
	return nil
}

func (x *SignalingEvent) GetRtcIceCandidateInit() *RTCIceCandidateInit {
	if x, ok := x.GetEvent().(*SignalingEvent_RtcIceCandidateInit); ok {
		return x.RtcIceCandidateInit
	}
	return nil
}

type isSignalingEvent_Event interface {
	isSignalingEvent_Event()
}

type SignalingEvent_RtcIceServer struct {
	RtcIceServer *RTCIceServer `protobuf:"bytes,1,opt,name=rtc_ice_server,json=rtcIceServer,proto3,oneof"`
}

type SignalingEvent_SessionDescription struct {
	SessionDescription *SessionDescription `protobuf:"bytes,2,opt,name=session_description,json=sessionDescription,proto3,oneof"`
}

type SignalingEvent_RtcIceCandidateInit struct {
	RtcIceCandidateInit *RTCIceCandidateInit `protobuf:"bytes,3,opt,name=rtc_ice_candidate_init,json=rtcIceCandidateInit,proto3,oneof"`
}

func (*SignalingEvent_RtcIceServer) isSignalingEvent_Event() {}

func (*SignalingEvent_SessionDescription) isSignalingEvent_Event() {}

func (*SignalingEvent_RtcIceCandidateInit) isSignalingEvent_Event() {}

// https://developer.mozilla.org/en-US/docs/Web/API/RTCIceServer
//
// for shared STUN server between Client and Server
// could add additional fields for TURN server
type RTCIceServer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Urls []string `protobuf:"bytes,1,rep,name=urls,proto3" json:"urls,omitempty"`
}

func (x *RTCIceServer) Reset() {
	*x = RTCIceServer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_signaling_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RTCIceServer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RTCIceServer) ProtoMessage() {}

func (x *RTCIceServer) ProtoReflect() protoreflect.Message {
	mi := &file_proto_signaling_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RTCIceServer.ProtoReflect.Descriptor instead.
func (*RTCIceServer) Descriptor() ([]byte, []int) {
	return file_proto_signaling_proto_rawDescGZIP(), []int{1}
}

func (x *RTCIceServer) GetUrls() []string {
	if x != nil {
		return x.Urls
	}
	return nil
}

// https://developer.mozilla.org/en-US/docs/Web/API/RTCSessionDescription
type SessionDescription struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type SessionDescription_SDPType `protobuf:"varint,1,opt,name=type,proto3,enum=SessionDescription_SDPType" json:"type,omitempty"`
	// Follows the format specified here: https://tools.ietf.org/html/rfc4566#section-5
	Sdp string `protobuf:"bytes,2,opt,name=sdp,proto3" json:"sdp,omitempty"`
}

func (x *SessionDescription) Reset() {
	*x = SessionDescription{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_signaling_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SessionDescription) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SessionDescription) ProtoMessage() {}

func (x *SessionDescription) ProtoReflect() protoreflect.Message {
	mi := &file_proto_signaling_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SessionDescription.ProtoReflect.Descriptor instead.
func (*SessionDescription) Descriptor() ([]byte, []int) {
	return file_proto_signaling_proto_rawDescGZIP(), []int{2}
}

func (x *SessionDescription) GetType() SessionDescription_SDPType {
	if x != nil {
		return x.Type
	}
	return SessionDescription_SDP_TYPE_UNSPECIFIED
}

func (x *SessionDescription) GetSdp() string {
	if x != nil {
		return x.Sdp
	}
	return ""
}

// https://developer.mozilla.org/en-US/docs/Web/API/RTCIceCandidateInit
type RTCIceCandidateInit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// https://developer.mozilla.org/en-US/docs/Web/API/RTCIceCandidateInit/candidate
	Candidate        string `protobuf:"bytes,1,opt,name=candidate,proto3" json:"candidate,omitempty"`
	SdpMid           string `protobuf:"bytes,2,opt,name=sdp_mid,json=sdpMid,proto3" json:"sdp_mid,omitempty"`
	SdpMLineIndex    uint32 `protobuf:"varint,3,opt,name=sdp_m_line_index,json=sdpMLineIndex,proto3" json:"sdp_m_line_index,omitempty"`
	UsernameFragment string `protobuf:"bytes,4,opt,name=username_fragment,json=usernameFragment,proto3" json:"username_fragment,omitempty"`
}

func (x *RTCIceCandidateInit) Reset() {
	*x = RTCIceCandidateInit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_signaling_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RTCIceCandidateInit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RTCIceCandidateInit) ProtoMessage() {}

func (x *RTCIceCandidateInit) ProtoReflect() protoreflect.Message {
	mi := &file_proto_signaling_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RTCIceCandidateInit.ProtoReflect.Descriptor instead.
func (*RTCIceCandidateInit) Descriptor() ([]byte, []int) {
	return file_proto_signaling_proto_rawDescGZIP(), []int{3}
}

func (x *RTCIceCandidateInit) GetCandidate() string {
	if x != nil {
		return x.Candidate
	}
	return ""
}

func (x *RTCIceCandidateInit) GetSdpMid() string {
	if x != nil {
		return x.SdpMid
	}
	return ""
}

func (x *RTCIceCandidateInit) GetSdpMLineIndex() uint32 {
	if x != nil {
		return x.SdpMLineIndex
	}
	return 0
}

func (x *RTCIceCandidateInit) GetUsernameFragment() string {
	if x != nil {
		return x.UsernameFragment
	}
	return ""
}

var File_proto_signaling_proto protoreflect.FileDescriptor

var file_proto_signaling_proto_rawDesc = []byte{
	0x0a, 0x15, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x69, 0x6e,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xe5, 0x01, 0x0a, 0x0e, 0x53, 0x69, 0x67, 0x6e,
	0x61, 0x6c, 0x69, 0x6e, 0x67, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x35, 0x0a, 0x0e, 0x72, 0x74,
	0x63, 0x5f, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x52, 0x54, 0x43, 0x49, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x48, 0x00, 0x52, 0x0c, 0x72, 0x74, 0x63, 0x49, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x12, 0x46, 0x0a, 0x13, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13,
	0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x12, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x44, 0x65,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4b, 0x0a, 0x16, 0x72, 0x74, 0x63,
	0x5f, 0x69, 0x63, 0x65, 0x5f, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x69,
	0x6e, 0x69, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x52, 0x54, 0x43, 0x49,
	0x63, 0x65, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x69, 0x74, 0x48,
	0x00, 0x52, 0x13, 0x72, 0x74, 0x63, 0x49, 0x63, 0x65, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x49, 0x6e, 0x69, 0x74, 0x42, 0x07, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22,
	0x22, 0x0a, 0x0c, 0x52, 0x54, 0x43, 0x49, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12,
	0x12, 0x0a, 0x04, 0x75, 0x72, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x75,
	0x72, 0x6c, 0x73, 0x22, 0x8b, 0x02, 0x0a, 0x12, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x44,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2f, 0x0a, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1b, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x53, 0x44,
	0x50, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x73,
	0x64, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x64, 0x70, 0x22, 0xb1, 0x01,
	0x0a, 0x07, 0x53, 0x44, 0x50, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x14, 0x53, 0x44, 0x50,
	0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45,
	0x44, 0x10, 0x00, 0x12, 0x09, 0x0a, 0x05, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x10, 0x01, 0x12, 0x12,
	0x0a, 0x0e, 0x53, 0x44, 0x50, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4f, 0x46, 0x46, 0x45, 0x52,
	0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x70, 0x72, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72, 0x10, 0x02,
	0x12, 0x15, 0x0a, 0x11, 0x53, 0x44, 0x50, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x50, 0x52, 0x41,
	0x4e, 0x53, 0x57, 0x45, 0x52, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x61, 0x6e, 0x73, 0x77, 0x65,
	0x72, 0x10, 0x03, 0x12, 0x13, 0x0a, 0x0f, 0x53, 0x44, 0x50, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f,
	0x41, 0x4e, 0x53, 0x57, 0x45, 0x52, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08, 0x72, 0x6f, 0x6c, 0x6c,
	0x62, 0x61, 0x63, 0x6b, 0x10, 0x04, 0x12, 0x15, 0x0a, 0x11, 0x53, 0x44, 0x50, 0x5f, 0x54, 0x59,
	0x50, 0x45, 0x5f, 0x52, 0x4f, 0x4c, 0x4c, 0x42, 0x41, 0x43, 0x4b, 0x10, 0x04, 0x1a, 0x02, 0x10,
	0x01, 0x22, 0xa2, 0x01, 0x0a, 0x13, 0x52, 0x54, 0x43, 0x49, 0x63, 0x65, 0x43, 0x61, 0x6e, 0x64,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x69, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x61, 0x6e,
	0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x61,
	0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x73, 0x64, 0x70, 0x5f, 0x6d,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x64, 0x70, 0x4d, 0x69, 0x64,
	0x12, 0x27, 0x0a, 0x10, 0x73, 0x64, 0x70, 0x5f, 0x6d, 0x5f, 0x6c, 0x69, 0x6e, 0x65, 0x5f, 0x69,
	0x6e, 0x64, 0x65, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x0d, 0x73, 0x64, 0x70, 0x4d,
	0x4c, 0x69, 0x6e, 0x65, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x2b, 0x0a, 0x11, 0x75, 0x73, 0x65,
	0x72, 0x6e, 0x61, 0x6d, 0x65, 0x5f, 0x66, 0x72, 0x61, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x46, 0x72,
	0x61, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x42, 0x12, 0x5a, 0x10, 0x7a, 0x6f, 0x6f, 0x6d, 0x67, 0x61,
	0x6d, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_proto_signaling_proto_rawDescOnce sync.Once
	file_proto_signaling_proto_rawDescData = file_proto_signaling_proto_rawDesc
)

func file_proto_signaling_proto_rawDescGZIP() []byte {
	file_proto_signaling_proto_rawDescOnce.Do(func() {
		file_proto_signaling_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_signaling_proto_rawDescData)
	})
	return file_proto_signaling_proto_rawDescData
}

var file_proto_signaling_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_signaling_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_proto_signaling_proto_goTypes = []interface{}{
	(SessionDescription_SDPType)(0), // 0: SessionDescription.SDPType
	(*SignalingEvent)(nil),          // 1: SignalingEvent
	(*RTCIceServer)(nil),            // 2: RTCIceServer
	(*SessionDescription)(nil),      // 3: SessionDescription
	(*RTCIceCandidateInit)(nil),     // 4: RTCIceCandidateInit
}
var file_proto_signaling_proto_depIdxs = []int32{
	2, // 0: SignalingEvent.rtc_ice_server:type_name -> RTCIceServer
	3, // 1: SignalingEvent.session_description:type_name -> SessionDescription
	4, // 2: SignalingEvent.rtc_ice_candidate_init:type_name -> RTCIceCandidateInit
	0, // 3: SessionDescription.type:type_name -> SessionDescription.SDPType
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_proto_signaling_proto_init() }
func file_proto_signaling_proto_init() {
	if File_proto_signaling_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_signaling_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignalingEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_signaling_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RTCIceServer); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_signaling_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SessionDescription); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_signaling_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RTCIceCandidateInit); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_proto_signaling_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*SignalingEvent_RtcIceServer)(nil),
		(*SignalingEvent_SessionDescription)(nil),
		(*SignalingEvent_RtcIceCandidateInit)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_signaling_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_signaling_proto_goTypes,
		DependencyIndexes: file_proto_signaling_proto_depIdxs,
		EnumInfos:         file_proto_signaling_proto_enumTypes,
		MessageInfos:      file_proto_signaling_proto_msgTypes,
	}.Build()
	File_proto_signaling_proto = out.File
	file_proto_signaling_proto_rawDesc = nil
	file_proto_signaling_proto_goTypes = nil
	file_proto_signaling_proto_depIdxs = nil
}
