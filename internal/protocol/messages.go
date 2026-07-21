package protocol

type Payload interface {
	Type() string
}

type Login struct {
	Username string `json:"username"`
}

func (Login) Type() string {
	return TypeLogin
}

type Message struct {
	Text string `json:"text"`
}

func (Message) Type() string {
	return TypeMessage
}
