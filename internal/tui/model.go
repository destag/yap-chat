package tui

import (
	"encoding/json"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/destag/yap-chat/internal/protocol"
)

type Model struct {
	messages []string

	incoming <-chan protocol.Packet
}

func New(incoming <-chan protocol.Packet) Model {
	return Model{
		messages: []string{},
		incoming: incoming,
	}
}

func (m Model) Init() tea.Cmd {
	return waitForPacket(m.incoming)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case PacketMsg:
		var message protocol.Message

		err := json.Unmarshal(
			msg.Packet.Data,
			&message,
		)

		if err == nil {
			m.messages = append(
				m.messages,
				message.Text,
			)
		}

		return m, waitForPacket(m.incoming)
	}

	return m, nil
}

func (m Model) View() string {
	result := "yap\n\n"

	for _, message := range m.messages {
		result += message + "\n"
	}

	return result
}

func waitForPacket(ch <-chan protocol.Packet) tea.Cmd {
	return func() tea.Msg {
		packet, ok := <-ch

		if !ok {
			return ChannelClosedMsg{}
		}

		return PacketMsg{
			Packet: packet,
		}
	}
}

type PacketMsg struct {
	Packet protocol.Packet
}

type ChannelClosedMsg struct{}
