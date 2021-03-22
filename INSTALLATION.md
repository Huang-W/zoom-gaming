# ZoomGaming

#### gRPC - Protobuf

Installing Go ([Instructions](https://golang.org/doc/install))

Installing npm [Instructions](https://docs.npmjs.com/cli/v7/commands/npm-install)

Installing protobuf-compiler ([Link to source](https://grpc.io/docs/protoc-installation/))
- Linux, `apt install -y protobuf-compiler`
- MacOS, `brew install protobuf`

Installing Go plugins for protoc ([Link to source](https://grpc.io/docs/languages/go/quickstart/))
- Module mode should be on `export GO111MODULE=on` or `go env -w GO111MODULE=0`
- `go get google.golang.org/protobuf/cmd/protoc-gen-go`
- Make sure that protoc can find the plugins `export PATH="$PATH:$(go env GOPATH)/bin"`

Protobuf compile to **Go** ([Reference](https://developers.google.com/protocol-buffers/docs/reference/go-generated))
```shell
protoc \
    --go_out=game/ \
    --go_opt=paths=source_relative \
    proto/signaling.proto
```

Protobuf compile to **Javascript** with commonjs imports ([Reference](https://developers.google.com/protocol-buffers/docs/reference/javascript-generated#commonjs-imports))
```shell
protoc \
    --js_out=import_style=commonjs,binary:web/js/ \
    proto/signaling.proto
```
