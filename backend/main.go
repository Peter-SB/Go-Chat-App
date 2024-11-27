package main

import (
	"go-chat-app/server"
	"log"
	"net/http"
)

// Main: The entry point focused on high-level setup.
func main() {
	server.InitDBConnection()
	http.HandleFunc("/history", server.GetChatHistoryHandler) // New endpoint for chat history

	http.HandleFunc("/ws", server.HandleConnections)

	http.HandleFunc("/register", server.Register)
	http.HandleFunc("/login", server.LoginUser)
	http.HandleFunc("/logout", server.LogoutUser)
	http.HandleFunc("/profile", server.Profile)

	// Launch background processes
	go server.StartBroadcastListener()
	go server.StartNotifyActiveUsers()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Run Command: `go run main.go`
