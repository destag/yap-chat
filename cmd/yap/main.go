package main

import (
	"net"

	"github.com/destag/yap-chat/internal/client"
	"github.com/destag/yap-chat/internal/transport"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		panic(err)
	}

	t := transport.NewTCPTransport(conn)

	client := client.New(t)

	client.Start()

	err = client.Login("alice")
	if err != nil {
		panic(err)
	}
}
