package server

import (
	"encoding/json"
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
		h.handleLogin(event)

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

	if client.username != "" {
		h.broadcastMessage(
			fmt.Sprintf("%s left the chat", client.username),
		)
	}
}

func (h *Hub) handleLogin(event HubEvent) {
	if event.Client.username != "" {
		return
	}

	var login protocol.Login

	err := json.Unmarshal(event.Packet.Data, &login)
	if err != nil {
		return
	}

	if login.Username == "" {
		return
	}

	if h.usernameTaken(login.Username) {
		// TODO: send error to client
		return
	}

	event.Client.username = login.Username

	h.broadcastMessage(
		fmt.Sprintf("%s joined the chat", login.Username),
	)
}

func (h *Hub) usernameTaken(username string) bool {
	for client := range h.clients {
		if client.username == username {
			return true
		}
	}

	return false
}

func (h *Hub) broadcastMessage(text string) {
	packet, err := protocol.New(protocol.Message{Text: text})
	if err != nil {
		return
	}

	h.broadcast(packet)
}
