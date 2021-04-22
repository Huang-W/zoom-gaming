package game

import (
	_ "context"
	"time"

	proto "google.golang.org/protobuf/proto"

	pb "zoomgaming/proto"
)

func ExampleInputs() {

	g, _ := NewGame(TestGame)
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
