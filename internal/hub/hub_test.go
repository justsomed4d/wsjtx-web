package hub

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestEnqueueDropOldest(t *testing.T) {
	c := &Client{send: make(chan BroadcastMessage, 2)}
	c.enqueue(BroadcastMessage{Type: websocket.TextMessage, Payload: []byte("a")})
	c.enqueue(BroadcastMessage{Type: websocket.TextMessage, Payload: []byte("b")})
	c.enqueue(BroadcastMessage{Type: websocket.TextMessage, Payload: []byte("c")})

	if len(c.send) != 2 {
		t.Fatalf("len(send)=%d want 2", len(c.send))
	}
	first := <-c.send
	second := <-c.send
	if string(first.Payload) != "b" || string(second.Payload) != "c" {
		t.Fatalf("got %q then %q, want b then c", first.Payload, second.Payload)
	}
}

func TestBroadcastNonBlocking(t *testing.T) {
	h := New()
	for i := 0; i < broadcastSize; i++ {
		h.Broadcast(BroadcastMessage{Type: websocket.TextMessage, Payload: []byte("x")})
	}
	done := make(chan struct{})
	go func() {
		h.Broadcast(BroadcastMessage{Type: websocket.TextMessage, Payload: []byte("drop")})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked when queue was full")
	}
}
