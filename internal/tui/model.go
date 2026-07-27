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
		cmds = append(cmds, m.handleKey(msg))

	case tea.WindowSizeMsg:
		cmds = append(cmds, m.handleResize(msg))

	case PacketMsg:
		cmds = append(cmds, m.handlePacket(msg))
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

func (m *Model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {

	case "ctrl+c":
		return m.quit()

	case "enter":
		return m.submitInput()
	}

	return nil
}

func (m *Model) submitInput() tea.Cmd {
	text := strings.TrimSpace(m.input.Value())

	if text == "" {
		return nil
	}

	m.input.Reset()

	if strings.HasPrefix(text, "/") {
		return m.handleCommand(
			strings.Fields(text[1:]),
		)
	}

	packet, err := protocol.New(protocol.SendMessage{
		Text: text,
	})
	if err != nil {
		return nil
	}

	_ = m.client.Send(packet)

	return nil
}

func (m *Model) quit() tea.Cmd {
	return func() tea.Msg {
		m.client.Close()
		return tea.Quit()
	}
}

func (m *Model) who() tea.Cmd {
	packet, err := protocol.New(protocol.WhoRequest{})
	if err != nil {
		return nil
	}

	_ = m.client.Send(packet)

	return nil
}

func (m *Model) handleCommand(args []string) tea.Cmd {
	if len(args) == 0 {
		return nil
	}

	switch args[0] {

	case "who":
		return m.who()

	case "help":
		m.messages = append(
			m.messages,
			"* Available commands: /help /quit /who",
		)
		m.refreshViewport()

	case "quit":
		return m.quit()

	default:
		m.messages = append(
			m.messages,
			"* Unknown command: "+args[0],
		)
		m.refreshViewport()
	}

	return nil
}

func (m *Model) handlePacket(msg PacketMsg) tea.Cmd {
	switch msg.Packet.Type {

	case protocol.TypeChatMessage:
		message, err := protocol.DecodePayload[protocol.ChatMessage](msg.Packet)
		if err != nil {
			return nil
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
			return nil
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

	case protocol.TypeLoginResponse:
		response, err := protocol.DecodePayload[protocol.LoginResponse](msg.Packet)
		if err != nil {
			return nil
		}

		if !response.Success {
			m.messages = append(
				m.messages,
				"* Login failed: "+response.Error,
			)
		}

	case protocol.TypeWhoResponse:
		response, err := protocol.DecodePayload[protocol.WhoResponse](msg.Packet)
		if err != nil {
			return nil
		}

		m.messages = append(
			m.messages,
			"* Connected users:",
		)

		for _, username := range response.Users {
			m.messages = append(
				m.messages,
				"  - "+username,
			)
		}

	}

	m.refreshViewport()
	return waitForPacket(m.client.Incoming())
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.viewport.SetWidth(msg.Width)
	m.viewport.SetHeight(msg.Height - 4)
	return nil
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

type PacketMsg struct {
	Packet protocol.Packet
}

type ChannelClosedMsg struct{}
