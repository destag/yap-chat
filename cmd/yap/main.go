package main

import (
	"fmt"
	"net"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/destag/yap-chat/internal/client"
	"github.com/destag/yap-chat/internal/transport"
	"github.com/destag/yap-chat/internal/tui"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		panic(err)
	}

	t := transport.NewTCPTransport(conn)

	client := client.New(t)

	client.Start()

	program := tea.NewProgram(
		tui.New(client.Incoming()),
	)

	go func() {
		err := client.Login("alice")
		if err != nil {
			fmt.Println(err)
		}
	}()

	_, err = program.Run()
	if err != nil {
		panic(err)
	}
}
