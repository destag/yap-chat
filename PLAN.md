# LAN Chat - Development Plan

## Overview

A local network terminal chat application written in Go.

The application consists of:

- `yapd` - server daemon
- `yap` - terminal UI client

Features:

- single global chat room
- multiple users on the same LAN
- real-time messaging
- persistent message history
- terminal-based UI
- lightweight deployment

The server is the source of truth. Clients connect, send commands, and receive events.

---

# Architecture

                TCP
                 |
    +------------+------------+
    |                         |
 chat client              chat client
    |                         |
    +------------+------------+
                 |
                 |
              chatd
                 |
          +--------------+
          |     Hub      |
          +--------------+
          | clients      |
          | broadcasts   |
          | events       |
          +--------------+
                 |
          +--------------+
          |   Storage    |
          +--------------+
          | SQLite       |
          +--------------+


---

# Repository Structure

chat/
├── cmd/
│ ├── chat/
│ │ └── main.go
│ │
│ └── chatd/
│ └── main.go
│
├── internal/
│ ├── protocol/
│ │ ├── packet.go
│ │ ├── message.go
│ │ └── events.go
│ │
│ ├── transport/
│ │ ├── transport.go
│ │ └── tcp.go
│ │
│ ├── server/
│ │ ├── server.go
│ │ ├── hub.go
│ │ ├── client.go
│ │ └── handlers.go
│ │
│ ├── client/
│ │ └── client.go
│ │
│ ├── storage/
│ │ ├── sqlite.go
│ │ └── migrations.go
│ │
│ └── tui/
│ ├── model.go
│ ├── update.go
│ ├── view.go
│ └── styles.go
│
├── migrations/
│ └── 001_initial.sql
│
├── docs/
│ └── protocol.md
│
├── PLAN.md
├── README.md
├── go.mod
└── go.sum


---

# External Dependencies

## TUI

### Bubble Tea

github.com/charmbracelet/bubbletea


Purpose:

- application event loop
- keyboard handling
- rendering

---

### Bubbles

github.com/charmbracelet/bubbles

Purpose:

- textarea
- viewport
- lists
- reusable TUI components

---

### Lip Gloss


github.com/charmbracelet/lipgloss


Purpose:

- colors
- layout styling
- borders

---

## Database

### SQLite


modernc.org/sqlite


Purpose:

- embedded database
- no CGO
- single binary deployment

---

## Standard Library

The following will use Go standard library:

- networking (`net`)
- JSON protocol (`encoding/json`)
- logging (`log/slog`)
- synchronization (`sync`)
- configuration (`flag` initially)

---

# Core Design

## Hub

The Hub is the central component.

Responsibilities:

- track connected clients
- broadcast messages
- manage online users
- distribute events

Example:

```go
type Hub struct {
    clients map[*Client]bool

    register chan *Client
    unregister chan *Client
    broadcast chan Message
}

```

The Hub owns all shared state.

Clients never modify shared state directly.

---

Protocol

Initial transport:

TCP

Encoding:

newline-delimited JSON

Example:

```json
{
  "type": "message",
  "data": {
    "text": "hello"
  }
}
```

Message types:

```
login
message
history
user_joined
user_left
users
ping
pong
error
```

---

## Database

SQLite schema:

messages

```
id
created_at
username
text
```

Stored:
- chat history
- system events

Not stored:
- online users
- connections

---

# Implementation Tasks

## Phase 1 - Project foundation
- [x] Create Go module
- [ ] Create repository structure
- [ ] Add dependencies
- [ ] Add basic README
- [ ] Add Makefile

## Phase 2 - Protocol
- [x] Define packet format
- [x] Define message types
- [x] Implement JSON encoding
- [x] Implement JSON decoding
- [ ] Add protocol tests

## Phase 3 - Transport
- [ ] Create transport interface
- [ ] Implement TCP transport
- [ ] Implement read loop
- [ ] Implement write loop
- [ ] Add connection handling

## Phase 4 - Server
- [ ] Create TCP listener
- [ ] Create Hub
- [ ] Implement client registration
- [ ] Implement client removal
- [ ] Implement broadcasting
- [ ] Implement username handling
- [ ] Implement graceful shutdown

## Phase 5 - Storage
 Add SQLite initialization
 Add migrations
 Save messages
 Load recent history
 Add storage tests

## Phase 6 - Client Library
 Connect to server
 Authenticate username
 Send messages
 Receive events
 Handle reconnect

## Phase 7 - TUI
 Create Bubble Tea application
 Add chat viewport
 Add input field
 Render messages
 Display users
 Add timestamps
 Add scrolling
 Add keyboard shortcuts

## Phase 8 - Quality of Life
 Configuration file
 Server discovery via mDNS
 Better error handling
 Logging
 Tests
 Documentation

## Future Ideas
WebSocket transport
browser client
authentication
encryption
private messages
file transfer
reactions
message editing
multiple rooms
