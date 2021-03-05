package utils

import (
	"testing"

	pb "zoomgaming/signaling"

	webrtc "github.com/pion/webrtc/v3"

	proto "google.golang.org/protobuf/proto"
)

// Make sure that the proto-generated types can be converted into pion types
//
// webrtc.ICEServer - pb.RTCIceServer
func TestRTCIceServerTypeMatch(t *testing.T) {

	var (
		s1  pb.RTCIceServer
		s2  webrtc.ICEServer
		s3  pb.RTCIceServer
		err error
	)

	s1 = pb.RTCIceServer{
		Urls: []string{"stun:stun.l.google.com:19302"},
	}

	err = ConvertFromProtoMessage(s1.ProtoReflect(), &s2)
	if err != nil {
		t.Errorf("Converting from s1 to s2 was unsuccessful - %s", err.Error())
	}

	err = ConvertToProtoMessage(&s2, s3.ProtoReflect())
	if err != nil {
		t.Errorf("Converting from s2 to s3 was unsuccessful - %s", err.Error())
	}

	// check equality
	if !proto.Equal(s1.ProtoReflect().Interface(), s3.ProtoReflect().Interface()) {
		t.Errorf("Want equality between s1 and s3")
	}
}

// The sdp string field follows the format specified here: https://tools.ietf.org/html/rfc4566#section-5
//
// webrtc.SessionDescription - pb.SessionDescription
func TestRTCSessionDescriptionTypeMatch(t *testing.T) {

	var (
		s1  pb.SessionDescription
		s2  webrtc.SessionDescription
		s3  pb.SessionDescription
		err error
	)

	s1 = pb.SessionDescription{
		Type: 1,
		Sdp: `v=0
o=alice 2890844526 2890844526 IN IP4 host.anywhere.com
s=
c=IN IP4 host.anywhere.com
t=0 0
m=audio 49170 RTP/AVP 0
a=rtpmap:0 PCMU/8000
m=video 51372 RTP/AVP 31
a=rtpmap:31 H261/90000
m=video 53000 RTP/AVP 32
a=rtpmap:32 MPV/90000`,
	}

	err = ConvertFromProtoMessage(s1.ProtoReflect(), &s2)
	if err != nil {
		t.Errorf("Converting from s1 to s2 was unsuccessful - %s", err.Error())
	}

	err = ConvertToProtoMessage(&s2, s3.ProtoReflect())
	if err != nil {
		t.Errorf("Converting from s2 to s3 was unsuccessful - %s", err.Error())
	}

	// check equality
	if !proto.Equal(s1.ProtoReflect().Interface(), s3.ProtoReflect().Interface()) {
		t.Errorf("Want equality between s1 and s3")
	}
}

func TestRTCIceCandidateTypeMatch(t *testing.T) {

	var (
		s1  pb.RTCIceCandidateInit
		s2  webrtc.ICECandidateInit
		s3  pb.RTCIceCandidateInit
		err error
	)

	s1 = pb.RTCIceCandidateInit{
		Candidate:        "candidate:4234997325 1 udp 2043278322 192.168.0.56 44323 typ host",
		SdpMid:           "video",
		SdpMLineIndex:    14,
		UsernameFragment: "CsxzEWmoKpJyscFj",
	}

	err = ConvertFromProtoMessage(s1.ProtoReflect(), &s2)
	if err != nil {
		t.Errorf("Converting from s1 to s2 was unsuccessful - %s", err.Error())
	}

	err = ConvertToProtoMessage(&s2, s3.ProtoReflect())
	if err != nil {
		t.Errorf("Converting from s2 to s3 was unsuccessful - %s", err.Error())
	}

	// check equality
	if !proto.Equal(s1.ProtoReflect().Interface(), s3.ProtoReflect().Interface()) {
		t.Errorf("Want equality between s1 and s3")
	}
}
