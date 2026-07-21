package transport

import (
	"bufio"
	"context"
	"net"

	"github.com/destag/yap-chat/internal/protocol"
)

type TCPTransport struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func NewTCPTransport(conn net.Conn) *TCPTransport {
	return &TCPTransport{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
}

func (t *TCPTransport) Read(ctx context.Context) (protocol.Packet, error) {
	if err := t.setDeadline(ctx); err != nil {
		return protocol.Packet{}, err
	}

	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return protocol.Packet{}, err
	}

	return protocol.Decode(line)
}

func (t *TCPTransport) Write(ctx context.Context, packet protocol.Packet) error {
	if err := t.setDeadline(ctx); err != nil {
		return err
	}

	data, err := protocol.Encode(packet)
	if err != nil {
		return err
	}

	if _, err := t.writer.Write(append(data, '\n')); err != nil {
		return err
	}

	return t.writer.Flush()
}

func (t *TCPTransport) Close() error {
	return t.conn.Close()
}

func (t *TCPTransport) setDeadline(ctx context.Context) error {
	deadline, ok := ctx.Deadline()

	if !ok {
		return nil
	}

	return t.conn.SetDeadline(deadline)
}
