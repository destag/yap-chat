package server

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/destag/yap-chat/internal/protocol"
	"github.com/destag/yap-chat/internal/transport"
)

type Client struct {
	transport transport.Transport
	outgoing  chan protocol.Packet
	done      chan struct{}
	closeOnce sync.Once

	username string

	// hub *Hub
}

func NewClient(t transport.Transport) *Client {
	return &Client{
		transport: t,
		outgoing:  make(chan protocol.Packet),
		done:      make(chan struct{}),
	}
}

func (c *Client) Start() {
	go c.readLoop()
	go c.writeLoop()
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.transport.Close()
		close(c.outgoing)
	})
}

func (c *Client) Send(packet protocol.Packet) error {
	select {
	case c.outgoing <- packet:
		return nil

	case <-c.done:
		return errors.New("client closed")
	}
}

func (c *Client) writeLoop() {
	defer c.Close()

	for {
		select {

		case packet, ok := <-c.outgoing:
			if !ok {
				return
			}

			err := c.transport.Write(
				context.Background(),
				packet,
			)
			if err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

func (c *Client) readLoop() {
	defer c.Close()

	for {
		packet, err := c.transport.Read(context.Background())
		if err != nil {
			c.disconnect()
			return
		}

		c.handlePacket(packet)
	}
}

func (c *Client) handlePacket(packet protocol.Packet) {
	fmt.Println(packet.Type)
}

func (c *Client) disconnect() {}
