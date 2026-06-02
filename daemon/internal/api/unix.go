package api

import (
	"log"
	"net"
	"os"
)

func StartUnixServer(sockPath string) {
	os.Remove(sockPath)
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Printf("[unix] error: %v", err)
		return
	}
	defer ln.Close()
	log.Printf("[unix] ready at %s", sockPath)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	cmd := string(buf[:n])
	switch cmd {
	case "ping":
		conn.Write([]byte("pong"))
	case "status":
		conn.Write([]byte("running"))
	default:
		conn.Write([]byte("ok"))
	}
}