package server

import (
	"fmt"
	"time"

	"github.com/destag/yap-chat/internal/protocol"
)

const (
	// Keep the last 100 messages. Prune in batches once
	// the history reaches 120 entries to avoid pruning
	// on every append.
	maxHistory          = 100
	historyPruneTrigger = 120
)

type HubEvent struct {
	Client *Client
	Packet protocol.Packet
}

type Hub struct {
	clients map[*Client]bool
	history []protocol.ChatMessage

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

	case protocol.TypeSendMessage:
		send, err := protocol.Unpack[protocol.SendMessage](event.Packet)
		if err != nil {
			return
		}

		msg := protocol.ChatMessage{
			Author:    event.Client.username,
			Text:      send.Text,
			Timestamp: time.Now().UTC(),
		}

		h.addHistory(msg)

		packet := protocol.MustPack(msg)
		h.broadcast(packet)

	case protocol.TypeLoginRequest:
		h.handleLogin(event)

	case protocol.TypeWhoRequest:
		h.handleWho(event)

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

func (h *Hub) handleWho(event HubEvent) {
	users := []string{}

	for client := range h.clients {
		users = append(users, client.username)
	}

	packet := protocol.MustPack(protocol.WhoResponse{
		Users: users,
	})

	event.Client.Send(packet)
}

func (h *Hub) handleLogin(event HubEvent) {
	if event.Client.username != "" {
		return
	}

	login, err := protocol.Unpack[protocol.LoginRequest](event.Packet)
	if err != nil {
		return
	}

	if login.Username == "" {
		packet := protocol.MustPack(protocol.LoginResponse{
			Success: false,
			Error:   "username cannot be empty",
		})

		event.Client.Send(packet)
		return
	}

	if h.usernameTaken(login.Username) {
		packet := protocol.MustPack(protocol.LoginResponse{
			Success: false,
			Error:   "username already taken",
		})
		event.Client.Send(packet)

		return
	}

	event.Client.username = login.Username

	packet := protocol.MustPack(protocol.LoginResponse{
		Success: true,
	})
	event.Client.Send(packet)

	h.broadcastMessage(
		fmt.Sprintf("%s joined the chat", login.Username),
	)

	packet = protocol.MustPack(protocol.HistoryResponse{
		Messages: h.history,
	})
	event.Client.Send(packet)
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
	packet := protocol.MustPack(protocol.SystemMessage{
		Text:      text,
		Timestamp: time.Now().UTC(),
	})

	h.broadcast(packet)
}

func (h *Hub) addHistory(msg protocol.ChatMessage) {
	h.history = append(h.history, msg)

	if len(h.history) <= historyPruneTrigger {
		return
	}

	h.history = h.history[len(h.history)-maxHistory:]
}
