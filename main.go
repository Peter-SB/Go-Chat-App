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

	go client.writePump()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, client)
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

func handleMessages() {
	for {
		msg := <-broadcast

		// Marshal message to JSON
		jsonMessage, _ := json.Marshal(msg)

		for client := range clients {
			select {
			case client.Send <- jsonMessage:
			default:
				close(client.Send)
				delete(clients, client)
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
