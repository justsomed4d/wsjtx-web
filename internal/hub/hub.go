package hub

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	sendBufferSize = 64
	broadcastSize  = 256
)

// BroadcastMessage is a framed WebSocket payload to send to all clients.
type BroadcastMessage struct {
	Type    int
	Payload []byte
}

// Client is one WebSocket peer with a dedicated outbound queue.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan BroadcastMessage
}

// Hub tracks WebSocket clients and fans out broadcast messages without blocking
// on slow peers.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
}

// New creates a Hub ready for Run and Register.
func New() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan BroadcastMessage, broadcastSize),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// NewClient wraps a connection for hub registration.
func NewClient(h *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:  h,
		conn: conn,
		send: make(chan BroadcastMessage, sendBufferSize),
	}
}

// Conn returns the underlying WebSocket connection.
func (c *Client) Conn() *websocket.Conn {
	return c.conn
}

// PrepareRead arms read deadlines and pong handling for keepalive.
func (c *Client) PrepareRead() {
	c.conn.SetReadLimit(1 << 20) // 1 MiB
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
}

// WritePump sends queued messages and periodic pings until the client drops.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(msg.Type, msg.Payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) enqueue(msg BroadcastMessage) {
	select {
	case c.send <- msg:
		return
	default:
	}
	// Buffer full: drop oldest, then try once more so a slow client cannot stall others.
	select {
	case <-c.send:
	default:
	}
	select {
	case c.send <- msg:
	default:
		log.Printf("dropping message for slow websocket client")
	}
}

// Run processes register/unregister/broadcast until the process exits.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				client.enqueue(message)
			}
		}
	}
}

// Broadcast enqueues a message for all connected clients. Drops if the hub queue is full.
func (h *Hub) Broadcast(msg BroadcastMessage) {
	select {
	case h.broadcast <- msg:
	default:
		log.Printf("broadcast queue full, dropping message")
	}
}

// Register adds a client connection.
func (h *Hub) Register(c *Client) {
	h.register <- c
}

// Unregister removes a client connection and closes its send queue.
func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}
