package client

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

	outgoing chan protocol.Packet
	incoming chan protocol.Packet

	done      chan struct{}
	closeOnce sync.Once

	username string
}

func New(t transport.Transport) *Client {
	return &Client{
		transport: t,
		outgoing:  make(chan protocol.Packet),
		incoming:  make(chan protocol.Packet),
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
	})
}

func (c *Client) Send(packet protocol.Packet) error {
	select {
	case c.outgoing <- packet:
		log.Printf("client queue outgoing packet: %+v", packet)
		return nil

	case <-c.done:
		return errors.New("client closed")
	}
}

func (c *Client) Login(username string) error {
	packet := protocol.MustPack(protocol.LoginRequest{
		Username: username,
	})

	c.username = username

	return c.Send(packet)
}

func (c *Client) Incoming() <-chan protocol.Packet {
	return c.incoming
}

func (c *Client) readLoop() {
	defer c.Close()

	for {
		packet, err := c.transport.Read(context.Background())
		if err != nil {
			return
		}

		select {
		case c.incoming <- packet:
			log.Printf("client queue incoming packet: %+v", packet)

		case <-c.done:
			return
		}
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

			log.Printf("writing packet: %+v", packet)

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
