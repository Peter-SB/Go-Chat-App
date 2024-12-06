package utils

import (
	"go-chat-app/models"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	clients       = make(map[*models.Client]bool)
	broadcast     = make(chan models.Message)
	notifyClients = make(chan struct{})
	mutex         sync.Mutex
)

// GetBroadcastChannel returns the broadcast channel.
func GetBroadcastChannel() chan models.Message {
	return broadcast
}

// GetNotifyClientsChannel returns the notifyClients channel.
func GetNotifyClientsChannel() chan struct{} {
	return notifyClients
}

// GetClients returns a reference to the clients map with the mutex.
func GetClients() (map[*models.Client]bool, *sync.Mutex) {
	return clients, &mutex
}

// MakeClient does the setup of the client object such as name, id, etc.
func MakeClient(r *http.Request, ws *websocket.Conn, user *models.User) *models.Client {
	displayName := user.Username
	if displayName == "" {
		displayName = "Anonymous"
	}

	client := &models.Client{
		ID:          uuid.New().String(),
		DisplayName: displayName,
		Conn:        ws,
		Send:        make(chan []byte),
	}
	return client
}

// RegisterClient adds a client to the active client pool.
func RegisterClient(client *models.Client) {
	mutex.Lock()
	defer mutex.Unlock()
	clients[client] = true
	notifyClients <- struct{}{}
}

// DeregisterClient removes a client from the active client pool.
func DeregisterClient(client *models.Client) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(clients, client)
	notifyClients <- struct{}{}
}

// CollectActiveUsers returns a list of display names of active clients.
func CollectActiveUsers() []string {
	mutex.Lock()
	defer mutex.Unlock()
	users := []string{}
	for client := range clients {
		users = append(users, client.DisplayName)
	}
	return users
}
