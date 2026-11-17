package server

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"

	"wsjtx-web/internal/audio"
	"wsjtx-web/internal/hub"
	"wsjtx-web/pkg/wsjtx"

	"github.com/gorilla/websocket"
)

// AddrStore remembers the last WSJT-X UDP peer for outbound commands.
type AddrStore struct {
	mu   sync.Mutex
	addr *net.UDPAddr
}

// Set records the latest remote UDP address.
func (s *AddrStore) Set(addr *net.UDPAddr) {
	s.mu.Lock()
	s.addr = addr
	s.mu.Unlock()
}

// Get returns the latest remote UDP address, or nil if none yet.
func (s *AddrStore) Get() *net.UDPAddr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

// Config holds dependencies for the HTTP/WebSocket server.
type Config struct {
	Addr     string
	StaticFS fs.FS
	Hub      *hub.Hub
	UDPConn  *net.UDPConn
	Addrs    *AddrStore
	Audio    *audio.Manager // nil when audio is disabled
	Origins  []string       // allowed WebSocket Origin values; "*" allows all
}

// Start registers routes and begins ListenAndServe (blocking).
func Start(cfg Config) error {
	upgrader := websocket.Upgrader{
		CheckOrigin: CheckOriginFunc(cfg.Origins),
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(cfg.StaticFS)))
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		cfg.handleWS(upgrader, w, r)
	})
	log.Printf("Starting HTTP server on %s (WS origins: %v)...", cfg.Addr, cfg.Origins)
	return http.ListenAndServe(cfg.Addr, mux)
}

func (cfg Config) handleWS(upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := hub.NewClient(cfg.Hub, ws)
	cfg.Hub.Register(client)
	go client.WritePump()
	defer cfg.Hub.Unregister(client)

	client.PrepareRead()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var cmd struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Printf("Error parsing command: %v", err)
			continue
		}

		switch cmd.Type {
		case wsjtx.CmdGetAudioDevices:
			cfg.handleGetAudioDevices()
			continue
		case wsjtx.CmdSetAudioDevice:
			cfg.handleSetAudioDevice(cmd.Data)
			continue
		}

		targetAddr := cfg.Addrs.Get()
		if targetAddr == nil {
			log.Println("No WSJT-X instance seen yet, cannot send command")
			continue
		}

		msg, ok := buildUDPMessage(cmd.Type, cmd.Data)
		if !ok {
			continue
		}

		var buf bytes.Buffer
		encoder := wsjtx.NewEncoder(&buf)
		if err := encoder.Encode(msg); err != nil {
			log.Printf("Error encoding message: %v", err)
			continue
		}

		if _, err := cfg.UDPConn.WriteToUDP(buf.Bytes(), targetAddr); err != nil {
			log.Printf("Error sending UDP: %v", err)
		} else {
			log.Printf("Sent %s to %s", cmd.Type, targetAddr)
		}
	}
}

func (cfg Config) handleGetAudioDevices() {
	devices := []wsjtx.AudioDevice{}
	if cfg.Audio == nil {
		log.Println("Audio disabled; returning empty audioDevices")
	} else {
		listed, err := cfg.Audio.ListDevices()
		if err != nil {
			log.Printf("Error listing devices: %v", err)
			return
		}
		devices = listed
	}

	resp, err := wsjtx.MarshalEnvelope(wsjtx.TypeAudioDevices, devices)
	if err != nil {
		log.Printf("Error marshaling audioDevices: %v", err)
		return
	}
	cfg.Hub.Broadcast(hub.BroadcastMessage{
		Type:    websocket.TextMessage,
		Payload: resp,
	})
}

func (cfg Config) handleSetAudioDevice(data json.RawMessage) {
	if cfg.Audio == nil {
		log.Println("Audio disabled; ignoring setAudioDevice")
		return
	}
	var d wsjtx.SetAudioDeviceCommand
	if err := json.Unmarshal(data, &d); err != nil {
		log.Printf("Error parsing setAudioDevice data: %v", err)
		return
	}
	if err := cfg.Audio.Start(d.ID); err != nil {
		log.Printf("Error setting audio device: %v", err)
	}
}

func buildUDPMessage(cmdType string, data json.RawMessage) (wsjtx.Message, bool) {
	switch cmdType {
	case wsjtx.CmdReply:
		var d wsjtx.ReplyCommand
		if err := json.Unmarshal(data, &d); err != nil {
			log.Printf("Error parsing reply data: %v", err)
			return nil, false
		}
		return &wsjtx.ReplyMessage{
			BaseMessage: wsjtx.BaseMessage{Id: "WSJTX-WEB"},
			Time:        d.Time,
			SNR:         d.SNR,
			DeltaTime:   d.DeltaTime,
			DeltaFreq:   d.DeltaFrequency,
			Mode:        d.Mode,
			Message:     d.Message,
			LowConf:     d.LowConfidence,
			Modifiers:   d.Modifiers,
		}, true
	case wsjtx.CmdHalt:
		var d wsjtx.HaltCommand
		if err := json.Unmarshal(data, &d); err != nil {
			log.Printf("Error parsing halt data: %v", err)
			return nil, false
		}
		return &wsjtx.HaltTxMessage{
			BaseMessage: wsjtx.BaseMessage{Id: "WSJTX-WEB"},
			AutoTxOnly:  d.AutoTxOnly,
		}, true
	default:
		log.Printf("Unknown command type: %s", cmdType)
		return nil, false
	}
}
