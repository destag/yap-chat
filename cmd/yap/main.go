package main

import (
	"flag"
	"fmt"
	"net"

	tea "charm.land/bubbletea/v2"

	"github.com/destag/yap-chat/internal/client"
	"github.com/destag/yap-chat/internal/transport"
	"github.com/destag/yap-chat/internal/tui"
)

func main() {
	username := flag.String(
		"name",
		"",
		"username",
	)

	flag.Parse()

	if *username == "" {
		fmt.Println("username is required")
		return
	}

	f, err := tea.LogToFile("debug-"+*username+".log", "debug")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		panic(err)
	}

	t := transport.NewTCPTransport(conn)

	client := client.New(t)

	client.Start()

	program := tea.NewProgram(
		tui.New(client),
	)

	go func() {
		err := client.Login(*username)
		if err != nil {
			fmt.Println(err)
		}
	}()

	_, err = program.Run()
	if err != nil {
		panic(err)
	}
}
