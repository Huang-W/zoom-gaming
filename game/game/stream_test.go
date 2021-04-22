package game

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestTestVP8(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, Port, 5004)

	s, err := NewStream(ctx, TestVP8)
	if err != nil {
		t.Errorf("error creating new stream")
	}

	var receivedPackets [][]byte = make([][]byte, 0)
	mu := &sync.Mutex{}
	s.Start()

	updates := s.Updates()
	go func() {
		for ch := range updates {
			go func() {
				for pckt := range ch {
					mu.Lock()
					receivedPackets = append(receivedPackets, pckt)
					mu.Unlock()
				}
			}()
		}
	}()

	time.Sleep(5 * time.Second)

	cancel()

	mu.Lock()
	defer mu.Unlock()
	if len(receivedPackets) < 50 {
		t.Errorf("Received %d packets after 5 seconds", len(receivedPackets))
	} else {
		t.Logf("Received %d packets after 5 seconds", len(receivedPackets))
	}
}

func TestTestOpus(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, Port, 4004)

	s, err := NewStream(ctx, TestOpus)
	if err != nil {
		t.Errorf("error creating new stream")
	}

	var receivedPackets [][]byte = make([][]byte, 0)
	mu := &sync.Mutex{}
	s.Start()

	updates := s.Updates()
	go func() {
		for ch := range updates {
			go func() {
				for pckt := range ch {
					mu.Lock()
					receivedPackets = append(receivedPackets, pckt)
					mu.Unlock()
				}
			}()
		}
	}()

	time.Sleep(5 * time.Second)

	cancel()

	mu.Lock()
	defer mu.Unlock()
	if len(receivedPackets) < 50 {
		t.Errorf("Received %d packets after 5 seconds", len(receivedPackets))
	} else {
		t.Logf("Received %d packets after 5 seconds", len(receivedPackets))
	}
}
