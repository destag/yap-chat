package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/destag/yap-chat/internal/protocol"
)

func (m Model) View() tea.View {
	v := tea.NewView(
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.headerView(),
			m.chatView(),
			m.inputView(),
			m.footerView(),
		),
	)

	v.AltScreen = true

	return v
}

func (m Model) headerView() string {
	return lipgloss.NewStyle().
		Bold(true).
		MarginBottom(1).
		Render("yap - " + m.server)
}

func (m Model) chatView() string {
	return m.viewport.View()
}

func (m Model) inputView() string {
	if m.loginFailed {
		return ""
	}

	return lipgloss.NewStyle().
		PaddingTop(1).
		PaddingBottom(1).
		Render(m.input.View())
}

func (m Model) footerView() string {
	if m.loginFailed {
		return "Press any key to exit"
	}

	return "/help for available commands"
}

func formatSystemMessage(msg protocol.SystemMessage) string {
	return fmt.Sprintf(
		"[%s] * %s",
		msg.Timestamp.Local().Format("15:04"),
		msg.Text,
	)
}

func formatChatMessage(msg protocol.ChatMessage) string {
	return fmt.Sprintf(
		"[%s] %s: %s",
		msg.Timestamp.Local().Format("15:04"),
		msg.Author,
		msg.Text,
	)
}
