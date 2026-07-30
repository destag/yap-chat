package server

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/destag/yap-chat/internal/protocol"
	"github.com/destag/yap-chat/internal/transport"
)

type Client struct {
	transport transport.Transport
	hub       *Hub

	outgoing  chan protocol.Packet
	done      chan struct{}
	closeOnce sync.Once

	username string
}

func NewClient(t transport.Transport, hub *Hub) *Client {
	return &Client{
		transport: t,
		hub:       hub,
		outgoing:  make(chan protocol.Packet),
		done:      make(chan struct{}),
	}
}

func (c *Client) Start() {
	c.hub.register <- c

	go c.readLoop()
	go c.writeLoop()
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.hub.unregister <- c
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

			log.Printf("sending packet %s: %d\n", packet.Type, len(packet.Data))

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
			c.Close()
			return
		}

		log.Printf("receiving packet %s: %d\n", packet.Type, len(packet.Data))

		select {
		case c.hub.events <- HubEvent{
			Client: c,
			Packet: packet,
		}:

		case <-c.done:
			return
		}
	}
}
