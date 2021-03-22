var pb = require('./proto/signaling_pb');

var iceServer = new pb.RTCIceServer();
iceServer.setUrlsList(["stun:stun.l.google.com:19302"]);

var sdp = new pb.SessionDescription();
sdp.setType(pb.SessionDescription.SDPType['SDP_TYPE_OFFER']);
sdp.setSdp(`v=0
o=alice 2890844526 2890844526 IN IP4 host.anywhere.com
s=
c=IN IP4 host.anywhere.com
t=0 0
m=audio 49170 RTP/AVP 0
a=rtpmap:0 PCMU/8000
m=video 51372 RTP/AVP 31
a=rtpmap:31 H261/90000
m=video 53000 RTP/AVP 32
a=rtpmap:32 MPV/90000`);

var iceCand = new pb.RTCIceCandidateInit();
iceCand.setCandidate("candidate:4234997325 1 udp 2043278322 192.168.0.56 44323 typ host");

var evt1 = new pb.SignalingEvent();
// Checks which event is set
// Possible choices:
//
// proto.SignalingEvent.EventCase = {
//   EVENT_NOT_SET: 0,
//   RTC_ICE_SERVER: 1,
//   SESSION_DESCRIPTION: 2,
//   RTC_ICE_CANDIDATE_INIT: 3
// };
//
// https://developers.google.com/protocol-buffers/docs/reference/javascript-generated#oneof

// 0
console.log(evt1.getEventCase());
evt1.setRtcIceServer(iceServer);
// 1
console.log(evt1.getEventCase());

var byte_array = evt1.serializeBinary();
var evt2 = new pb.SignalingEvent.deserializeBinary(byte_array);

console.log(evt1.toObject());
console.log(evt2.toObject());

console.log(evt1.toObject() == evt2.toObject());
