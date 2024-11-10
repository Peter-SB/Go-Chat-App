package server

import (
	"encoding/json"
	"log"
)

// Broadcasting logic file.

// StartBroadcastListener listens for chat messages on the broadcast channel and sends them to all connected clients.
func StartBroadcastListener() {
	for msg := range broadcast {
		messageBytes, _ := json.Marshal(msg)
		mutex.Lock()

		for client := range clients {
			select {
			case client.Send <- messageBytes:
			default:
				// Remove client if unresponsive
				DeregisterClient(client)
			}
		}
		mutex.Unlock()
	}
}

// StartNotifyActiveUsers listens for updates and notifies all clients of the current active user list.
func StartNotifyActiveUsers() {
	for range notifyClients {
		activeUsers := CollectActiveUsers()

		msg := ActiveUsersMessage{
			Type:  "activeUsers",
			Users: activeUsers,
		}

		messageBytes, _ := json.Marshal(msg)

		mutex.Lock()
		for client := range clients {
			select {
			case client.Send <- messageBytes:
			default:
				// Remove unresponsive client
				DeregisterClient(client)
			}
		}
		mutex.Unlock()
	}
}

// BroadcastMessage sends a message to the broadcast channel when a user sends a chat message.
func BroadcastMessage(msg Message) {
	// Save to database
	err := SaveMessage(msg)
	if err != nil {
		log.Printf("Failed to save message to DB: %v", err)
	}

	// Broadcast to all connected clients
	broadcast <- msg
}
