package transport

import (
	"context"

	"github.com/destag/yap-chat/internal/protocol"
)

type Transport interface {
	Read(ctx context.Context) (protocol.Packet, error)
	Write(ctx context.Context, packet protocol.Packet) error
	Close() error
}
