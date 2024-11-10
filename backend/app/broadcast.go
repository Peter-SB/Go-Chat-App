package app

import (
	"encoding/json"
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

// BroadcastMessage sends a message to the broadcast channel.
func BroadcastMessage(msg Message) {
	broadcast <- msg
}
