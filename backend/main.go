package main

import (
	"go-chat-app/server"
	"log"
	"net/http"
)

// Main: The entry point focused on high-level setup.
func main() {
	server.InitDBConnection()

	server.SetupRoutes()

	// Launch background processes
	go server.StartBroadcastListener()
	go server.StartNotifyActiveUsers()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Run Command: `go run main.go`
