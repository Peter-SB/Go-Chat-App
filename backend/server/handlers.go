package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocket handlers focus on establishing connections and adding clients to the pool.

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow any origin. Todo: adjust in production for security.
		return true
	},
}

// HandleConnections handles when a user connects. It upgrades the HTTP connection to a WebSocket connection,
// adds the user to the client map, starts listening for messages from the client, and reads incoming websocket messages
func HandleConnections(w http.ResponseWriter, r *http.Request) {

	// Upgrade the HTTP connection to WebSocket.
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	// Create a new Client instance and adds it to the clients map
	client := MakeClient(r, ws)
	RegisterClient(client)

	// Start listening for messages from this client
	go handleClientMessages(client)

	// Read incoming websocket messages
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			DeregisterClient(client)
			break
		}
		BroadcastMessage(msg)
	}
}

func handleClientMessages(client *Client) {
	defer DeregisterClient(client)
	for {
		msg := <-client.Send
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("write error:", err)
			return
		}
	}
}

// GetChatHistoryHandler gets the users chat history from the db
func GetChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	messages, err := GetChatHistory()
	if err != nil {
		http.Error(w, "Failed to retrieve chat history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*") // backend and frontend are on different domains or ports, add CORS headers to the backend. Todo: investigate further and fix security issues.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
