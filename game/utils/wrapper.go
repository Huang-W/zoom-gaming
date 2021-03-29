package utils

import (
	webrtc "github.com/pion/webrtc/v3"

	pb "zoomgaming/proto"
)

func WrapRTCIceServer(iceServer *webrtc.ICEServer) (pb.SignalingEvent, error) {

	var iceServer_pb pb.RTCIceServer // convert from pion/webrtc to protobuf
	if err := ConvertToProtoMessage(iceServer, iceServer_pb.ProtoReflect()); err != nil {
		return pb.SignalingEvent{}, err
	}

	return pb.SignalingEvent{
		Event: &pb.SignalingEvent_RtcIceServer{
			RtcIceServer: &iceServer_pb,
		},
	}, nil
}

func WrapSessionDescription(sdp *webrtc.SessionDescription) (pb.SignalingEvent, error) {

	var sdp_pb pb.SessionDescription // convert from pion/webrtc to protobuf
	if err := ConvertToProtoMessage(sdp, sdp_pb.ProtoReflect()); err != nil {
		return pb.SignalingEvent{}, err
	}

	return pb.SignalingEvent{
		Event: &pb.SignalingEvent_SessionDescription{
			SessionDescription: &sdp_pb,
		},
	}, nil
}

func WrapRTCIceCandidateInit(iceCandInit *webrtc.ICECandidateInit) (pb.SignalingEvent, error) {

	var cand pb.RTCIceCandidateInit
	if err := ConvertToProtoMessage(iceCandInit, cand.ProtoReflect()); err != nil {
		return pb.SignalingEvent{}, err
	}

	return pb.SignalingEvent{
		Event: &pb.SignalingEvent_RtcIceCandidateInit{
			RtcIceCandidateInit: &cand,
		},
	}, nil
}
