package api

import (
	"log"
	"net"
	"os"
)

func StartUnixServer(sockPath string, handler func(string) string) {
	os.Remove(sockPath)
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Printf("[unix] listen error: %v", err)
		return
	}
	defer ln.Close()
	log.Printf("[unix] listening on %s", sockPath)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 1024)
			n, _ := c.Read(buf)
			cmd := string(buf[:n])
			resp := handler(cmd)
			c.Write([]byte(resp))
		}(conn)
	}
}