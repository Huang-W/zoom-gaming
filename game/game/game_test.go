package game

import (
	_ "context"
	"sync"
	"time"

	"github.com/pion/rtp"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
)

func ExampleInputs() {

	pool := &sync.Pool{
		New: func() interface{} {
			return new(rtp.Packet)
		},
	}

	g, _ := NewGame(TestGame, pool)
	defer g.Close()

	input1 := make(chan proto.Message)
	input2 := make(chan proto.Message)
	input3 := make(chan proto.Message)
	input4 := make(chan proto.Message)

	g.AttachInputStream(input1)
	g.AttachInputStream(input2)
	g.AttachInputStream(input3)
	g.AttachInputStream(input4)

	input1 <- &pb.Echo{}
	input2 <- &pb.Echo{}
	input3 <- &pb.Echo{}
	input4 <- &pb.Echo{}

	time.Sleep(1 * time.Second)
	// Output:
	//
	//
	//
	//
}
