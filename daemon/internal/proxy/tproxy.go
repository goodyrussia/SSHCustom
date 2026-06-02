package proxy

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"
)

func StartTPROXY(ctx context.Context, port int, client interface{}) {
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		log.Printf("[tproxy] listen error: %v", err)
		return
	}
	defer ln.Close()
	log.Printf("[tproxy] listening on :%d", port)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			go handleTPROXY(conn)
		}
	}
}

func handleTPROXY(conn net.Conn) {
	defer conn.Close()
	// Placeholder - real TPROXY + SSH forwarding logic goes here
	time.Sleep(100 * time.Millisecond)
}