package server

import (
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	clients       = make(map[*Client]bool)
	broadcast     = make(chan Message)
	notifyClients = make(chan struct{})
	mutex         sync.Mutex // Protects the clients map
)

// MakeClient does the setup of the client object such as name, id, etc.
func MakeClient(r *http.Request, ws *websocket.Conn) *Client {
	displayName := r.URL.Query().Get("displayName")
	if displayName == "" {
		displayName = "Anonymous"
	}

	client := &Client{
		ID:          uuid.New().String(),
		DisplayName: displayName,
		Conn:        ws,
		Send:        make(chan []byte),
	}
	return client
}

// RegisterClient adds a client to the active client pool.
func RegisterClient(client *Client) {
	mutex.Lock()
	defer mutex.Unlock()
	clients[client] = true
	notifyClients <- struct{}{}
}

// DeregisterClient removes a client from the active client pool.
func DeregisterClient(client *Client) {
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
