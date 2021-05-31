package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	pb "tt/proto"
)

func main() {
	fmt.Printf("|%15s | %10s | %15s | %10s | %15s | %21s | %40s | \n",
    "no of tickers","json"," gzipped json"," proto", "gzipped proto", "proto size(%) of json","gzipped proto size(%) of gzipped json")
	for _, dataSize := range []int{0, 1, 2, 10, 20, 200, 2000, 20000} {
		eventArr := createTestDatata(dataSize)
		jsonl, gzJsonlen, protol, gzProto := jsonProtoLengts(eventArr)
		fmt.Printf("|%15d | %10d | %15d | %10d | %15d | %21f | %40f | \n", dataSize, jsonl, gzJsonlen, protol, gzProto, float32(gzProto)/float32(gzJsonlen), float32(protol)/float32(jsonl))
	}
}

func createTestDatata(numberOfEvents int) []pb.KeyPressEvent {

	events := make([]pb.KeyPressEvent, numberOfEvents)

	for i := 0; i < numberOfEvents; i++ {
		events[i] = pb.KeyPressEvent{
      Direction: pb.KeyPressEvent_Direction(rand.Int31n(2)+1),
      Key: pb.KeyPressEvent_Key(rand.Int31n(8)+1),
		}
	}

	return events
}

func jsonProtoLengts(eventArr []pb.KeyPressEvent) (jsonLen, gzipJSONLen, protoLen, gzpProto int) {

  for _, event := range eventArr {
    data, _ := proto.Marshal(&event)
    protoLen += len(data)
    jsonified, _ := json.Marshal(&event)
    jsonLen += len(jsonified)
    gzipJSONLen += gzipLen(jsonified)
    gzpProto += gzipLen(data)
  }

	return
}

func gzipLen(jsonData []byte) int {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(jsonData); err != nil {
		panic(err)
	}
	if err := gz.Flush(); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}

	return b.Len()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
