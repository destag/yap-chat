package protocol

import "encoding/json"

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

func DecodePayload[T any](packet Packet) (T, error) {
	var payload T

	err := json.Unmarshal(packet.Data, &payload)

	return payload, err
}
