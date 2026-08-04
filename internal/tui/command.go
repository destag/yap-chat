package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/destag/yap-chat/internal/protocol"
)

type CommandFunc func(*Model, []string) tea.Cmd

type Command struct {
	Name        string
	Description string
	Run         CommandFunc
}

var (
	commandList []Command
	commands    map[string]Command
)

func init() {
	commandList = []Command{
		{"help", "Show this help", cmdHelp},
		{"who", "List connected users", cmdWho},
		{"quit", "Disconnect and exit", cmdQuit},
	}
	commands = make(map[string]Command, len(commandList))
	for _, c := range commandList {
		commands[c.Name] = c
	}
}

func cmdHelp(m *Model, args []string) tea.Cmd {
	m.messages = append(m.messages, "* Available commands:")

	for _, info := range commandList {
		m.messages = append(m.messages, fmt.Sprintf("  /%-5s %s", info.Name, info.Description))
	}

	m.refreshViewport()
	return nil
}

func cmdWho(m *Model, args []string) tea.Cmd {
	packet := protocol.MustPack(protocol.WhoRequest{})
	m.client.Send(packet)
	return nil
}

func cmdQuit(m *Model, args []string) tea.Cmd {
	return m.quit()
}

func completeCommand(value string) string {
	if !strings.HasPrefix(value, "/") {
		return value
	}

	partial := strings.TrimPrefix(value, "/")

	for _, c := range commandList {
		if strings.HasPrefix(c.Name, partial) {
			return "/" + c.Name
		}
	}

	return value
}

func (m *Model) handleCommand(args []string) tea.Cmd {
	if len(args) == 0 {
		return nil
	}

	cmd, ok := commands[args[0]]

	if !ok {
		m.messages = append(m.messages, "* Unknown command: "+args[0])
		m.refreshViewport()
		return nil
	}

	return cmd.Run(m, args[1:])
}
