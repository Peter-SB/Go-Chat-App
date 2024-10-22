package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var clients = make(map[*Client]bool)
var broadcast = make(chan Message)
var notifyClients = make(chan struct{}) // New channel to notify clients of changes

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	ID          string
	DisplayName string
	Conn        *websocket.Conn
	Send        chan []byte
}

type Message struct {
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ActiveUsersMessage represents the list of active users sent to all clients
type ActiveUsersMessage struct {
	Type  string   `json:"type"`  // "activeUsers"
	Users []string `json:"users"` // List of active display names
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ws.Close()

	id := uuid.New().String()
	displayName := r.URL.Query().Get("displayName")
	if displayName == "" {
		displayName = "Anonymous"
	}

	client := &Client{
		ID:          id,
		DisplayName: displayName,
		Conn:        ws,
		Send:        make(chan []byte),
	}
	clients[client] = true

	go notifyActiveUsers()
	go client.writePump()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, client)
			close(client.Send)
			go notifyActiveUsers()
			break
		}

		message := Message{
			Sender:    client.DisplayName,
			Content:   string(msg),
			Timestamp: time.Now(),
		}
		broadcast <- message
	}
}

func notifyActiveUsers() {
	notifyClients <- struct{}{} // Notify to trigger the active users broadcast
}

func handleMessages() {
	for {
		select {
		case msg := <-broadcast:
			// Marshal message to JSON and send to all clients
			jsonMessage, _ := json.Marshal(msg)
			for client := range clients {
				select {
				case client.Send <- jsonMessage:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}

		case <-notifyClients:
			// When notified, create the active users list and broadcast to all clients
			activeUsers := []string{}
			for client := range clients {
				activeUsers = append(activeUsers, client.DisplayName)
			}

			activeUsersMessage := ActiveUsersMessage{
				Type:  "activeUsers",
				Users: activeUsers,
			}
			jsonActiveUsers, _ := json.Marshal(activeUsersMessage)

			for client := range clients {
				select {
				case client.Send <- jsonActiveUsers:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}
		}
	}
}

func (client *Client) writePump() {
	for {
		msg := <-client.Send
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("write error:", err)
			return
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	go handleMessages()

	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
