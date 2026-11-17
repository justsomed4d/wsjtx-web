package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net"
	"os"

	"wsjtx-web/internal/audio"
	"wsjtx-web/internal/hub"
	"wsjtx-web/internal/server"
	"wsjtx-web/pkg/wsjtx"

	"github.com/gorilla/websocket"
)

//go:embed dist
var distFS embed.FS

const defaultOrigins = "http://localhost:5173,http://127.0.0.1:5173,http://localhost:8080,http://127.0.0.1:8080"

func main() {
	udpAddr := flag.String("udp", ":2237", "UDP address to receive WSJT-X traffic")
	httpAddr := flag.String("http", ":8080", "HTTP/WebSocket listen address")
	enableAudio := flag.Bool("audio", false, "Enable audio capture for waterfall (malgo/CGO)")
	originsDefault := defaultOrigins
	if v := os.Getenv("WSJTX_WEB_ORIGINS"); v != "" {
		originsDefault = v
	}
	origins := flag.String("origins", originsDefault, "Comma-separated allowed WebSocket Origin values; use * to allow all (env: WSJTX_WEB_ORIGINS)")
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", *udpAddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("listen udp %s: %v (is another app like GridTracker already bound to this port? use -udp to pick another)", addr, err)
	}
	defer conn.Close()

	log.Printf("Listening for WSJT-X on %s...", addr)

	h := hub.New()
	go h.Run()

	fsys, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatal(err)
	}

	addrs := &server.AddrStore{}

	var audioMgr *audio.Manager
	if *enableAudio {
		audioMgr, err = audio.New(h)
		if err != nil {
			log.Println("Failed to init audio manager:", err)
		} else if err := audioMgr.Start(""); err != nil {
			log.Println("Failed to start audio:", err)
		}
	} else {
		log.Println("Audio capture disabled (pass -audio to enable waterfall capture)")
	}

	go func() {
		if err := server.Start(server.Config{
			Addr:     *httpAddr,
			StaticFS: fsys,
			Hub:      h,
			UDPConn:  conn,
			Addrs:    addrs,
			Audio:    audioMgr,
			Origins:  server.ParseOrigins(*origins),
		}); err != nil {
			log.Fatal(err)
		}
	}()

	buf := make([]byte, 4096)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Error reading UDP:", err)
			continue
		}

		addrs.Set(remoteAddr)

		msg, err := wsjtx.ParsePacket(buf[:n])
		if err != nil {
			continue
		}

		wrappedJSON, err := wsjtx.MarshalEnvelope(wsjtx.TypeName(msg.Type()), msg)
		if err == nil {
			h.Broadcast(hub.BroadcastMessage{
				Type:    websocket.TextMessage,
				Payload: wrappedJSON,
			})
		}
	}
}
