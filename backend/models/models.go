package models

import (
	"time"

	"github.com/gorilla/websocket"
)

// Why: Consolidates all shared data structures into a single file, improving discoverability and cohesion.

// Client represents a WebSocket client .
type Client struct {
	ID          string
	DisplayName string
	Conn        *websocket.Conn
	Send        chan []byte
}

// Message represents a chat message.
type Message struct {
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ActiveUsersMessage represents the list of active users sent to all clients.
type ActiveUsersMessage struct {
	Type  string   `json:"type"`  // Always "activeUsers"
	Users []string `json:"users"` // List of active display names
}

type User struct {
	ID             int
	Username       string
	HashedPassword string
	SessionToken   string
	CSRFToken      string
}
