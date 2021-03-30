## Protobuf message classes

These are message classes used for communication between server and client.

The `SignalingEvent` message defined in `signaling.proto` is used in the WebSocket connection, and is one of three events: `RTCIceServer`, `SessionDescription`, or `RTCIceCandidateInit`.
- The `RTCIceServer` event is passed from server to client and is used as part of the configuration in the browser client's `RTCPeerConnection` constructor
- Both sides accept the `SessionDescription` message and use it to respectively `setRemoteDescription(session_description)`
- In a "balanced" bundle policy, there are three RTCDtlsTransport per connection, one for each type of track (video, audio, and data). Each transport has a pair of `RTCIceCandidateInit`, representing the two sides of a transport. One end of the connection is the controlling ICE agent (the offerer?) and will decide on which pair of ice candidates to use. Both sides should `addICECandidate(ice_cand_init)` when they receive this message.

### References
A brief explanation of ICE: https://webrtcforthecurious.com/docs/03-connecting/#ice
What is the Session Description Protocol?: https://webrtcforthecurious.com/docs/02-signaling/#what-is-the-session-description-protocol-sdp
