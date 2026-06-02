package proxy

import (
	"context"
	"log"
	"net"
	"strconv"
)

func StartDNSForwarder(ctx context.Context, cfg interface{}, client interface{}) {
	port := 53
	ln, err := net.ListenPacket("udp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		log.Printf("[dns] listen error: %v", err)
		return
	}
	defer ln.Close()
	log.Printf("[dns] forwarding on :%d", port)

	buf := make([]byte, 512)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, _, _ = ln.ReadFrom(buf)
			// TODO: forward DNS through SSH tunnel
		}
	}
}