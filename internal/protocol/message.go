package protocol

import (
	"time"
)

type Payload interface {
	Type() string
}

type Login struct {
	Username string `json:"username"`
}

func (Login) Type() string {
	return TypeLogin
}

type SendMessage struct {
	Text string `json:"text"`
}

func (SendMessage) Type() string {
	return TypeSendMessage
}

type ChatMessage struct {
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

func (ChatMessage) Type() string {
	return TypeChatMessage
}

type SystemMessage struct {
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

func (SystemMessage) Type() string {
	return TypeSystemMessage
}

type WhoRequest struct{}

func (WhoRequest) Type() string {
	return TypeWhoRequest
}

type WhoResponse struct {
	Users []string `json:"users"`
}

func (WhoResponse) Type() string {
	return TypeWhoResponse
}
