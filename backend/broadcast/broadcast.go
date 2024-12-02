package broadcast

import (
	"encoding/json"
	"log"

	"go-chat-app/db"
	"go-chat-app/models"
	"go-chat-app/utils"
)

// Broadcasting logic file.

// StartBroadcastListener listens for chat messages on the broadcast channel and sends them to all connected clients.
func StartBroadcastListener() {
	broadcast := utils.GetBroadcastChannel()
	clients, mutex := utils.GetClients()

	for msg := range broadcast {
		messageBytes, _ := json.Marshal(msg)
		mutex.Lock()

		for client := range clients {
			select {
			case client.Send <- messageBytes:
			default:
				// Remove client if unresponsive
				utils.DeregisterClient(client)
			}
		}
		mutex.Unlock()
	}
}

// StartNotifyActiveUsers listens for updates and notifies all clients of the current active user list.
func StartNotifyActiveUsers() {
	notifyClients := utils.GetNotifyClientsChannel()
	clients, mutex := utils.GetClients()

	for range notifyClients {
		activeUsers := utils.CollectActiveUsers()

		msg := models.ActiveUsersMessage{
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
				utils.DeregisterClient(client)
			}
		}
		mutex.Unlock()
	}
}

// BroadcastMessage sends a message to the broadcast channel when a user sends a chat message.
func BroadcastMessage(msg models.Message) {
	// Save to database
	err := db.SaveMessage(msg)
	if err != nil {
		log.Printf("Failed to save message to DB: %v", err)
	}

	// Broadcast to all connected clients
	broadcast := utils.GetBroadcastChannel()
	broadcast <- msg
}
