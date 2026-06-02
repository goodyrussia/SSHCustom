package api

import (
	"log"
	"net"
	"os"
)

var Version = "5.0.0"

func StartUnixServer(sockPath string) {
	os.Remove(sockPath)
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Printf("[unix] error: %v", err)
		return
	}
	defer ln.Close()
	log.Printf("[unix] listening on %s", sockPath)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)
	n, _ := conn.Read(buf)
	cmd := string(buf[:n])

	switch cmd {
	case "ping":
		conn.Write([]byte("pong"))
	case "status":
		conn.Write([]byte("running"))
	case "version":
		conn.Write([]byte(Version))
	default:
		conn.Write([]byte("ok"))
	}
}
