package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/destag/yap-chat/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	s := server.New(":9000")

	if err := s.ListenAndServe(ctx); err != nil {
		log.Fatal(err)
	}
}
