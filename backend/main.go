package main

import (
	"go-chat-app/app"
	"log"
	"net/http"
)

// Main: The entry point focused on high-level setup.
func main() {
	http.HandleFunc("/ws", app.HandleConnections)

	// Launch background processes
	go app.StartBroadcastListener()
	go app.StartNotifyActiveUsers()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Run Command: `go run main.go`
