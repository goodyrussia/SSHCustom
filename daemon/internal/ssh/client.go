package ssh

import (
	"log"
	"net"
	"time"
)

// Client represents SSH connection with payload support
type Client struct {
	conn net.Conn
	done chan struct{}
}

func Connect(cfg interface{}) (*Client, error) {
	// Placeholder - real SSH + payload injection here
	log.Println("[ssh] connected (placeholder)")
	c := &Client{done: make(chan struct{})}
	go func() {
		time.Sleep(60 * time.Second)
		close(c.done)
	}()
	return c, nil
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) Close() {
	close(c.done)
}