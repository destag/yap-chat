package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/destag/yap-chat/internal/client"
	"github.com/destag/yap-chat/internal/protocol"
)

type Model struct {
	client   *client.Client
	messages []string

	input textinput.Model
}

func New(client *client.Client) Model {
	input := textinput.New()

	input.Placeholder = "Type a message..."
	input.Focus()
	input.SetWidth(50)

	return Model{
		client:   client,
		messages: []string{},
		input:    input,
	}
}

func (m Model) Init() tea.Cmd {
	return waitForPacket(m.client.Incoming())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				break
			}

			packet, err := protocol.New(protocol.SendMessage{
				Text: text,
			})
			if err == nil {
				_ = m.client.Send(packet)
			}

			m.input.Reset()
		}

	case PacketMsg:
		switch msg.Packet.Type {

		case protocol.TypeChatMessage:
			message, err := protocol.DecodePayload[protocol.ChatMessage](msg.Packet)
			if err != nil {
				return m, nil
			}

			timestamp := message.Timestamp.
				Local().
				Format("15:04")

			msg := fmt.Sprintf(
				"[%s] %s: %s",
				timestamp,
				message.Author,
				message.Text,
			)

			m.messages = append(m.messages, msg)

		case protocol.TypeSystemMessage:
			message, err := protocol.DecodePayload[protocol.SystemMessage](msg.Packet)
			if err != nil {
				return m, nil
			}

			timestamp := message.Timestamp.
				Local().
				Format("15:04")

			msg := fmt.Sprintf(
				"[%s] * %s",
				timestamp,
				message.Text,
			)

			m.messages = append(m.messages, msg)
		}

		return m, waitForPacket(m.client.Incoming())
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	var b strings.Builder

	b.WriteString("yap\n\n")

	for _, msg := range m.messages {
		b.WriteString(msg)
		b.WriteByte('\n')
	}

	b.WriteString("\n")
	b.WriteString(m.input.View())

	return tea.NewView(b.String())
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
