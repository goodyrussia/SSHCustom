package ssh

import (
	"log"
	"net"
	"time"
)

// Client holds the SSH connection and payload state
type Client struct {
	conn   net.Conn
	done   chan struct{}
	host   string
	config interface{}
}

// Connect establishes SSH tunnel with payload injection
func Connect(cfg interface{}) (*Client, error) {
	// TODO: real SSH + payload injection using golang.org/x/crypto/ssh
	log.Println("[ssh] connecting with payload support...")

	c := &Client{
		done: make(chan struct{}),
	}

	// Placeholder connection
	go func() {
		time.Sleep(300 * time.Second)
		close(c.done)
	}()

	return c, nil
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) Close() {
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
}