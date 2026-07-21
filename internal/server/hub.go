package server

import "github.com/destag/yap-chat/internal/protocol"

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
			h.clients[client] = true

		case client := <-h.unregister:
			delete(h.clients, client)

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
		client.Send(packet)
	}
}
