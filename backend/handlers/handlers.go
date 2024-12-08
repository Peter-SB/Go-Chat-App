package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"go-chat-app/broadcast"
	"go-chat-app/models"
	"go-chat-app/services"
	"go-chat-app/utils"

	"github.com/gorilla/websocket"
)

// WebSocket handlers focus on establishing connections and adding clients to the pool.

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow any origin. Todo: adjust in production for security.
		return true
	},
}

// HandleConnections handles when a user connects. It authenticates, upgrades the HTTP connection to a WebSocket connection,
// adds the user to the client map, starts listening for messages from the client, and reads incoming websocket messages
func HandleConnections(services *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authenticate the user
		user, err := services.Auth.Authorise(r)
		if err != nil {
			log.Printf("Unauthorised WebSocket connection attempt: %v", err)
			http.Error(w, "Unauthorised", http.StatusUnauthorized)
			return
		}

		// Log the authorised user
		log.Printf("WebSocket connection authorised for user: %s", user.Username)

		// Upgrade the HTTP connection to WebSocket.
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer ws.Close()

		// Create a new Client instance and adds it to the clients map
		client := utils.MakeClient(r, ws, user)
		utils.RegisterClient(client)

		// Start listening for messages from this client
		go handleClientMessages(client)

		// Read incoming websocket messages
		for {
			var msg models.Message
			err := ws.ReadJSON(&msg)
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				utils.DeregisterClient(client)
				break
			}
			broadcast.BroadcastMessage(msg)
		}
	}
}

// handleClientMessages goroutine listening for messages from this client
func handleClientMessages(client *models.Client) {
	defer utils.DeregisterClient(client)
	for {
		msg := <-client.Send
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("write error:", err)
			return
		}
	}
}

// ChatHistoryHandler handles GET or DELETE requests for the chat history endpoint.
// todo: add paging
func ChatHistoryHandler(services *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			messages, err := services.DB.GetChatHistory()
			if err != nil {
				http.Error(w, "Failed to retrieve chat history", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(messages)

		case http.MethodDelete:
			err := services.DB.DeleteAllMessages()
			if err != nil {
				http.Error(w, "Failed to delete messages", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
