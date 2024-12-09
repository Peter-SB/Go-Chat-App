package models

import (
	"time"

	"github.com/gorilla/websocket"
)

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

// User represents a user in the db.
type User struct {
	ID             int
	Username       string
	HashedPassword string
	SessionToken   string
	CSRFToken      string
}

// ActiveUsersMessage represents the list of active users sent to all clients.
type ActiveUsersMessage struct {
	Type  string   `json:"type"`  // Always "activeUsers"
	Users []string `json:"users"` // List of active display names
}
