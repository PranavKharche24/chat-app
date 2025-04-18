package networking

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/PranavKharche24/chat-app/internal/encryption"
)

// Client wraps a TCP connection to the chat server.
type Client struct {
	conn     net.Conn
	Incoming chan string
	Errors   chan error
}

// NewClient dials the server at addr.
func NewClient(addr string) (*Client, error) {
	c := &Client{
		Incoming: make(chan string, 10),
		Errors:   make(chan error, 1),
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	c.conn = conn
	return c, nil
}

// Send sends a line (automatically adds "\n").
func (c *Client) Send(cmd string) error {
	_, err := fmt.Fprintf(c.conn, "%s\n", cmd)
	return err
}

// Receive starts reading server replies, decrypting CHAT: lines.
func (c *Client) Receive() {
	go func() {
		sc := bufio.NewScanner(c.conn)
		for sc.Scan() {
			line := sc.Text()
			if strings.HasPrefix(line, "CHAT:") {
				pt, err := encryption.Decrypt(strings.TrimPrefix(line, "CHAT:"))
				if err != nil {
					c.Incoming <- "[decrypt error]"
				} else {
					c.Incoming <- pt
				}
			} else {
				c.Incoming <- line
			}
		}
		if err := sc.Err(); err != nil {
			c.Errors <- err
		}
		close(c.Incoming)
	}()
}

// Close shuts the connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
