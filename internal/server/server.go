package server

import (
	"context"
	"net"

	"github.com/destag/yap-chat/internal/transport"
)

type Server struct {
	addr string
	hub  *Hub
}

func New(addr string) *Server {
	return &Server{
		addr: addr,
		hub:  NewHub(),
	}
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	go s.hub.Run()

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	defer listener.Close()

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return err
			}
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	t := transport.NewTCPTransport(conn)

	client := NewClient(t, s.hub)
	client.Start()
}
