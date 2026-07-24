package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/destag/yap-chat/internal/client"
	"github.com/destag/yap-chat/internal/protocol"
)

type Model struct {
	client *client.Client
	server string

	messages []string
	viewport viewport.Model

	input textinput.Model
}

func New(client *client.Client, server string) Model {
	input := textinput.New()

	input.Placeholder = "Type a message..."
	input.Focus()
	input.SetWidth(50)

	vp := viewport.New(
		viewport.WithWidth(80),
		viewport.WithHeight(20),
	)

	return Model{
		client:   client,
		server:   server,
		messages: []string{},
		viewport: vp,
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
			return m, quit(m.client)

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

	case tea.WindowSizeMsg:
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - 4)

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

			m.refreshViewport()

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

			m.refreshViewport()
		}

		return m, waitForPacket(m.client.Incoming())
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) refreshViewport() {
	m.viewport.SetContent(
		strings.Join(m.messages, "\n"),
	)

	m.viewport.GotoBottom()
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

func quit(client *client.Client) tea.Cmd {
	return func() tea.Msg {
		client.Close()
		return tea.Quit()
	}
}

type PacketMsg struct {
	Packet protocol.Packet
}

type ChannelClosedMsg struct{}
