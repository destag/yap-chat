package protocol

import (
	"time"
)

type Payload interface {
	Type() string
}

type LoginRequest struct {
	Username string `json:"username"`
}

func (LoginRequest) Type() string {
	return TypeLoginRequest
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (LoginResponse) Type() string {
	return TypeLoginResponse
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

type HistoryResponse struct {
	Messages []ChatMessage `json:"messages"`
}

func (HistoryResponse) Type() string {
	return TypeHistoryResponse
}
