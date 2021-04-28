package game

import (
	_ "bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"

	zutils "zoomgaming/utils"
)

/**

This represents a UDP stream with RTP packets
The constructor initializes the stream on a port using the caller's context

The Start() method must be called to begin reading from the stream.

No close method; the caller's context will end the stream.

*/

type Stream interface {
	Updates() chan (<-chan []byte) // only one channel of rtp packets is expected
}

type stream struct {
	listener *net.UDPConn
	receiver chan []byte
	cmd      *exec.Cmd
	updates  chan (<-chan []byte)
	ctx      context.Context
}

func NewStream(ctx context.Context, typ mediaStreamType) (s Stream, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%s", r))
			return
		}
	}()

	var cmd *exec.Cmd
	var listener *net.UDPConn

	port, ok := fromContext(ctx)
	if !ok {
		panic("Unable to extract port from context")
	}

	switch typ {
	case VideoSH:
		cmd = exec.CommandContext(ctx, "bash", "./video.sh", ":99", fmt.Sprintf("%d", port))
		break
	case AudioSH:
		cmd = exec.CommandContext(ctx, "bash", "./audio.sh", fmt.Sprintf("%d", port))
		break
	case TestH264:
		cmd = exec.CommandContext(ctx, "ffmpeg", "-re", "-f", "lavfi", "-i", "testsrc=size=640x480:rate=30",
			"-vcodec", "libx264", "-cpu-used", "5", "-deadline", "1", "-g", "10", "-error-resilient", "1", "-auto-alt-ref", "1", "-f", "rtp",
			fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", port))
		break
	case TestOpus:
		cmd = exec.CommandContext(ctx, "ffmpeg", "-f", "lavfi", "-i", "sine=frequency=1000",
			"-c:a", "libopus", "-b:a", "8000", "-sample_fmt", "s16p", "-ssrc", "1", "-payload_type", "111", "-f", "rtp", "-max_delay", "0", "-application", "lowdelay",
			fmt.Sprintf("rtp://127.0.0.1:%d?pkt_size=1200", port))
		break
	default:
		panic(fmt.Sprintf("Invalid MediaStreamType: %s", typ))
	}

	listener, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	zutils.FailOnError(err, "Error opening listener on port %d: ", port)

	sstream := &stream{
		listener: listener,
		cmd:      cmd,
		updates:  make(chan (<-chan []byte)),
		ctx:      ctx,
	}

	if err := sstream.start(); err != nil {
		panic(fmt.Sprintf("Error starting stream: %s", err))
	}

	s = sstream
	return
}

func (s *stream) Updates() chan (<-chan []byte) {
	return s.updates
}

func (s *stream) start() error {

	go s.readPackets()

	/** DEBUG
	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stdErr, err := s.cmd.StderrPipe()
	if err != nil {
		return err
	}
	*/
	if err := s.cmd.Start(); err != nil {
		return err
	}
	/** DEBUG
	go func() {
		buf := bufio.NewReader(stdout) // Notice that this is not in a loop
		for {
			line, _, _ := buf.ReadLine()
			if string(line) == "" {
				continue
			}
			log.Println(string(line))
		}
	}()

	go func() {
		buf := bufio.NewReader(stdErr) // Notice that this is not in a loop
		for {
			line, _, _ := buf.ReadLine()
			if string(line) == "" {
				continue
			}
			log.Println(string(line))
		}
	}()
	*/
	go func() {
		select {
		case <-s.ctx.Done():
			if err := s.listener.Close(); err != nil {
				log.Printf("Closing UDP listener: %s", err)
			}
		}
	}()

	return nil
}

func (s *stream) readPackets() {

	receiver := make(chan []byte, 400)
	s.receiver = receiver
	s.updates <- receiver

	defer func() {
		s.cmd.Wait()
		close(receiver)
		close(s.updates)
	}()

	// Read RTP packets forever and send them to the browser(s)
	for {
		inboundRTPPacket := make([]byte, 1600) // UDP MTU
		n, _, err := s.listener.ReadFrom(inboundRTPPacket)

		receiver <- inboundRTPPacket[:n]

		if err != nil {
			log.Printf("UDP Connection closed - exiting: %s", err)
			return
		}
	}
}
