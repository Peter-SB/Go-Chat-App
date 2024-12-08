package services

import (
	"go-chat-app/auth"
	"go-chat-app/db"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Services struct {
	DB   db.DBInterface
	Auth auth.AuthServiceInterface
}

// InitialiseServices initialises database and auth services
func InitialiseServices() (*db.MySQLDB, *Services) {
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

	services := &Services{
		DB:   mySQLDB,
		Auth: authService,
	}
	return mySQLDB, services
}
