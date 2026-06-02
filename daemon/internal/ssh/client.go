package ssh

import (
	"log"
	"net"
	"time"
)

// Client represents the SSH tunnel with payload support
type Client struct {
	conn net.Conn
	done chan struct{}
}

// Connect establishes the SSH connection with payload injection
func Connect(cfg interface{}) (*Client, error) {
	log.Println("[ssh] establishing tunnel with payload...")

	c := &Client{
		done: make(chan struct{}),
	}

	// Placeholder for real SSH + payload injection
	go func() {
		// In real implementation: connect via x/crypto/ssh + inject payload
		time.Sleep(300 * time.Second)
		close(c.done)
	}()

	return c, nil
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.done)
}