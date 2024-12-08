package main

import (
	"log"
	"net/http"

	"go-chat-app/broadcast"
	"go-chat-app/routes"
	"go-chat-app/services"
)

// main program entry point.
func main() {
	mySQLDB, services := services.InitialiseServices()

	// Inject dependencies for use by routes and broadcast listeners
	routes.SetupRoutes(services)
	broadcast.InitBroadcast(mySQLDB)

	// Launch background processes
	go broadcast.StartBroadcastListener()
	go broadcast.StartNotifyActiveUsers()

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Run Command: `go run main.go`
// Only Rebuild Backend Container Command: `docker-compose up --build backend`
