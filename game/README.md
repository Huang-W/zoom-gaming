### Game Server

#### Instructions

- [Install latest version of Go](https://golang.org/doc/install)
- `go env -w GO111MODULE=on`
- `go run main.go server.go`

Send HTTP upgrade requests to `ws://localhost:8080/ws` when the server is running.

Echo server running on endpoint. Any received Websocket messages that are in protobuf wireform will be echoed back to sender. Messages that are not serialized using protobuf will be logged to console.

#### Tests

- `go test -v zoomgaming/utils`
- `go test -v zoomgaming/websocket`
- `go test -race zoomgaming/websocket`
