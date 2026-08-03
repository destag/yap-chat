package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

var commands = []string{
	"who",
	"help",
	"quit",
}

func completeCommand(value string) string {
	if !strings.HasPrefix(value, "/") {
		return value
	}

	partial := strings.TrimPrefix(value, "/")

	for _, command := range commands {
		if strings.HasPrefix(command, partial) {
			return "/" + command
		}
	}

	return value
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
