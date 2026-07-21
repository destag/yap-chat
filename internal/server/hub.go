package server

import (
	"fmt"

	"github.com/destag/yap-chat/internal/protocol"
)

type HubEvent struct {
	Client *Client
	Packet protocol.Packet
}

type Hub struct {
	clients map[*Client]bool

	register   chan *Client
	unregister chan *Client
	events     chan HubEvent
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		events:     make(chan HubEvent),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.register:
			h.addClient(client)

		case client := <-h.unregister:
			h.removeClient(client)

		case event := <-h.events:
			h.handleEvent(event)
		}
	}
}

func (h *Hub) handleEvent(event HubEvent) {
	switch event.Packet.Type {

	case protocol.TypeMessage:
		h.broadcast(event.Packet)

	case protocol.TypeLogin:
	//

	default:
		// unknown packet
	}
}

func (h *Hub) broadcast(packet protocol.Packet) {
	for client := range h.clients {
		if err := client.Send(packet); err != nil {
			h.removeClient(client)
		}
	}
}

func (h *Hub) addClient(client *Client) {
	h.clients[client] = true

	fmt.Printf(
		"client connected (%d total)\n",
		len(h.clients),
	)
}

func (h *Hub) removeClient(client *Client) {
	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)

	fmt.Printf(
		"client disconnected (%d remaining)\n",
		len(h.clients),
	)
}
