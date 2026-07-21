package protocol

import "encoding/json"

const (
	TypeLogin      = "login"
	TypeMessage    = "message"
	TypeHistory    = "history"
	TypeUserJoined = "user_joined"
	TypeUserLeft   = "user_left"
	TypeUsers      = "users"
	TypePing       = "ping"
	TypePong       = "pong"
	TypeError      = "error"
)

type Packet struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func New(payload Payload) (Packet, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Packet{}, err
	}

	return Packet{
		Type: payload.Type(),
		Data: data,
	}, nil
}

func Encode(packet Packet) ([]byte, error) {
	return json.Marshal(packet)
}

func Decode(data []byte) (Packet, error) {
	var packet Packet

	err := json.Unmarshal(data, &packet)

	return packet, err
}
