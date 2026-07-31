package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	"github.com/destag/yap-chat/internal/client"
	"github.com/destag/yap-chat/internal/protocol"
)

const (
	minInputHeight = 1
	maxInputHeight = 5

	headerHeight = 2
	footerHeight = 3
)

type Model struct {
	client *client.Client
	server string

	loggedIn    bool
	loginFailed bool

	messages []string
	viewport viewport.Model

	input textarea.Model

	width  int
	height int
}

func New(client *client.Client, server string) Model {
	input := textarea.New()

	input.Placeholder = "Type a message..."
	input.Focus()
	input.SetWidth(50)
	input.SetHeight(minInputHeight)
	input.ShowLineNumbers = false
	input.DynamicHeight = true
	input.KeyMap.InsertNewline = key.NewBinding(
		key.WithKeys("alt+enter"),
	)

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

	return m, tea.Batch(cmds...)
}

func (m *Model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	if m.loginFailed {
		return m.quit()
	}

	switch msg.String() {
	case "ctrl+c", "ctrl+d":
		return m.quit()

	case "pgup":
		m.viewport.PageUp()

	case "pgdown":
		m.viewport.PageDown()

	case "ctrl+up":
		m.viewport.ScrollUp(1)

	case "ctrl+down":
		m.viewport.ScrollDown(1)

	case "enter":
		m.submitInput()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	m.updateInputHeight()

	return cmd
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

	packet := protocol.MustPack(protocol.SendMessage{
		Text: text,
	})

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
	packet := protocol.MustPack(protocol.WhoRequest{})

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
		message, err := protocol.Unpack[protocol.ChatMessage](msg.Packet)
		if err != nil {
			return nil
		}

		msg := formatChatMessage(message)
		m.messages = append(m.messages, msg)

	case protocol.TypeSystemMessage:
		message, err := protocol.Unpack[protocol.SystemMessage](msg.Packet)
		if err != nil {
			return nil
		}

		msg := formatSystemMessage(message)
		m.messages = append(m.messages, msg)

	case protocol.TypeLoginResponse:
		response, err := protocol.Unpack[protocol.LoginResponse](msg.Packet)
		if err != nil {
			return nil
		}

		if !response.Success {
			m.messages = append(
				m.messages,
				"* Login failed: "+response.Error,
			)

			m.refreshViewport()
			m.loginFailed = true

			return nil
		}

		m.loggedIn = true

	case protocol.TypeWhoResponse:
		response, err := protocol.Unpack[protocol.WhoResponse](msg.Packet)
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

	case protocol.TypeHistoryResponse:
		history, err := protocol.Unpack[protocol.HistoryResponse](msg.Packet)
		if err != nil {
			return nil
		}
		for _, msg := range history.Messages {
			m.messages = append(
				m.messages,
				formatChatMessage(msg),
			)
		}

	}

	m.refreshViewport()
	return waitForPacket(m.client.Incoming())
}

func (m *Model) handleResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.width = msg.Width
	m.height = msg.Height

	m.updateLayout()

	return nil
}

func (m *Model) refreshViewport() {
	m.viewport.SetContent(
		strings.Join(m.messages, "\n"),
	)

	m.viewport.GotoBottom()
}

func (m *Model) updateInputHeight() {
	if m.input.Height() > maxInputHeight {
		m.input.SetHeight(maxInputHeight)
	}

	m.updateLayout()
}

func (m *Model) updateLayout() {
	height := m.height - headerHeight - footerHeight - m.input.Height()

	m.viewport.SetHeight(height)
	m.viewport.SetWidth(m.width)
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
