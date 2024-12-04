package main

import (
	"log"
	"net/http"
	"os"

	"go-chat-app/auth"
	"go-chat-app/broadcast"
	"go-chat-app/db"
	"go-chat-app/routes"
	"go-chat-app/services"

	"github.com/joho/godotenv"
)

// main program entry point.
func main() {
	mySQLDB, services := initialiseServices()

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

// initialiseServices initialises database and auth services
func initialiseServices() (*db.MySQLDB, *services.Services) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Load environment variables
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	database := os.Getenv("DB_NAME")

	// Create the DSN
	dsn := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + database + "?parseTime=true" // parseTime=true option ensures that DATE, DATETIME, and TIMESTAMP types are scanned as time.Time in Go

	// Initialize the database
	mySQLDB, err := db.NewMySQLDB(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize the auth service
	authService := auth.NewAuthService(mySQLDB)

	services := &services.Services{
		DB:   mySQLDB,
		Auth: authService,
	}
	return mySQLDB, services
}

// Run Command: `go run main.go`
// Only Rebuild Backend Container Command: `docker-compose up --build backend`
