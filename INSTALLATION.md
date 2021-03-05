# ZoomGaming

#### gRPC - Protobuf

Installing Go ([Instructions](https://golang.org/doc/install))

Installing protobuf-compiler ([Link to source](https://grpc.io/docs/protoc-installation/))
- Linux, `apt install -y protobuf-compiler`
- MacOS, `brew install protobuf`

Installing Go plugins for protoc ([Link to source](https://grpc.io/docs/languages/go/quickstart/))
- Module mode should be on `export GO111MODULE=on` or `go env -w GO111MODULE=0`
- `go get google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc`
- Make sure that protoc can find the plugins `export PATH="$PATH:$(go env GOPATH)/bin"`

Installing grpc-web plugin for protoc ([Link to binary](https://github.com/grpc/grpc-web/releases/tag/1.2.1))
- Rename the binary: `mv protoc-gen-grpc-web-1.2.1-linux-x86_64 protoc-gen-grpc-web`
- Make the binary executable: `chmod +x protoc-gen-grpc-web`
- move the downloaded binary onto your $PATH: `sudo mv proto-gen-grpc-web /usr/local/bin` (or similar)

Generate protobuf message classes (Javascript and Go)
```shell
protoc \
    --js_out=import_style=commonjs:web/public/js/ \
    --go_out=game/ \
    --go_opt=paths=source_relative \
    --proto_path=proto/ \
    proto/signaling/signaling.proto proto/input/input.proto
```

#### Tests

- `go test -v zoomgaming/utils`
