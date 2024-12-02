package main

import (
	"log"
	"net/http"

	"go-chat-app/broadcast"
	"go-chat-app/db"
	"go-chat-app/routes"
)

// Main: The entry point focused on high-level setup.
func main() {
	db.InitDBConnection()

	routes.SetupRoutes()

	// Launch background processes
	go broadcast.StartBroadcastListener()
	go broadcast.StartNotifyActiveUsers()

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Run Command: `go run main.go`
// Only Rebuild Backend Container Command: `docker-compose up --build backend`
