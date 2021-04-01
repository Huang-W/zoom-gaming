package webrtc

import (
	"log"

	webrtc "github.com/pion/webrtc/v3"

	proto "google.golang.org/protobuf/proto"
	pref "google.golang.org/protobuf/reflect/protoreflect"
)

type DataChannel interface {
	Send(proto.Message) error
	Updates() chan (<-chan proto.Message)
	Close() error
}

type dataChannel struct {
	dc       *webrtc.DataChannel
	updates  chan (<-chan proto.Message)
	receiver chan proto.Message
	send     chan struct{}
}

func NewDataChannel(label DataChannelLabel, dc_impl *webrtc.DataChannel) DataChannel {

	dc := &dataChannel{
		dc:      dc_impl,
		updates: make(chan (<-chan proto.Message)),
		send:    make(chan struct{}),
	}

	dc_impl.OnOpen(func() {
		dc.receiver = make(chan proto.Message, 1024)
		dc.updates <- dc.receiver
		log.Printf("Data Channel '%s'-'%d' open.\n", dc_impl.Label(), *(dc_impl.ID()))
	})

	dc_impl.OnMessage(func(msg webrtc.DataChannelMessage) {

		b := msg.Data
		log.Printf("Message received on DataChannel '%s'", dc_impl.Label())

		var pb_msg pref.Message = mapping[label].New()
		err := proto.Unmarshal(b, pb_msg.Interface())
		if err != nil {
			return
		}

		// Send across receiving go channel
		dc.receiver <- pb_msg.Interface()
	})

	// Set bufferedAmountLowThreshold so that we can get notified when
	// we can send more
	dc_impl.SetBufferedAmountLowThreshold(bufferedAmountLowThreshold)

	// This callback is made when the current bufferedAmount becomes lower than the threadshold
	dc_impl.OnBufferedAmountLow(func() {

		dc.send <- struct{}{}

	})

	dc_impl.OnClose(func() {
		close(dc.receiver)
		log.Println("Data channel closed")
	})

	return dc
}

func (dc *dataChannel) Send(msg proto.Message) error {

	// marshal the message into a byte array
	b, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	// block if the size of the stream's buffer exceeds 1 MB
	if dc.dc.BufferedAmount()+uint64(len(b)) > maxBufferedAmount {
		<-dc.send
	}

	dc.dc.Send(b)
	return nil
}

func (dc *dataChannel) Updates() chan (<-chan proto.Message) {
	return dc.updates
}

func (dc *dataChannel) Close() error {

	err := dc.Close()
	return err
}
