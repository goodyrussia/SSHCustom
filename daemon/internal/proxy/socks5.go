package proxy

import (
	"context"
	"log"
	"net"
	"strconv"
)

func StartSOCKS5(ctx context.Context, port int, client interface{}) {
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		log.Printf("[socks5] listen error: %v", err)
		return
	}
	defer ln.Close()
	log.Printf("[socks5] listening on :%d", port)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, _ := ln.Accept()
			go handleSOCKS5(conn)
		}
	}
}

func handleSOCKS5(conn net.Conn) {
	defer conn.Close()
	// Placeholder for real SOCKS5
}