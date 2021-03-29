### Game Server

#### Instructions

- [Install latest version of Go](https://golang.org/doc/install)
- `go env -w GO111MODULE=on`
- `go run main.go server.go`

#### Description



#### Testing Browser

Test browser websocket code with the `ws://localhost:8080/ws` endpoint. Any messages with class pb.SignalingEvent will be echoed back to the sender.

Test browser webrtc code with the `ws://localhost:8080/webrtc` endpoint. Any messages on the "Echo" data channel will be echoed back to the sender. No support for media streams yet.

Received messages that are not serialized using protobuf will still be logged to console.

#### Tests

- `go test -v zoomgaming/utils`
- `go test -v -race zoomgaming/websocket``
- `go test -v -race zoomgaming/webrtc`
