package server

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

// InitDBConnection initializes the MySQL database connection.
func InitDBConnection() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get environment variables
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	database := os.Getenv("DB_NAME")

	// Create the DSN
	dsn := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + database + "?parseTime=true" // parseTime=true option ensures that DATE, DATETIME, and TIMESTAMP types are scanned as time.Time in Go

	// Connect to MySQL
	for i := 0; i < 10; i++ { // Retry up to 10 times
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to database after 10 attempts: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping MySQL: %v", err)
	}

	log.Println("Connected to MySQL database.")
}

// SaveMessage saves a chat message to the database.
func SaveMessage(msg Message) error {
	_, err := db.Exec(
		"INSERT INTO messages (sender, content, timestamp) VALUES (?, ?, ?)",
		msg.Sender, msg.Content, msg.Timestamp,
	)
	return err
}

// GetChatHistory retrieves historical messages from the database.
func GetChatHistory() ([]Message, error) {
	log.Println("Attempting to get chat history from MySQL database.")
	rows, err := db.Query("SELECT sender, content, timestamp FROM messages ORDER BY timestamp ASC")
	if err != nil {
		log.Printf("SQL error: %v", err)
		return nil, err
	}
	defer rows.Close()

	log.Println("MySQL db queried.")

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.Sender, &msg.Content, &msg.Timestamp)
		if err != nil {
			log.Printf("Row scan error: %v", err)
			log.Printf("Debugging row: sender=%v, content=%v, timestamp=%v", msg.Sender, msg.Content, msg.Timestamp)
			continue // Skip problematic rows
		}
		log.Printf("Retrieved message: %+v", msg)
		messages = append(messages, msg)
	}

	// Check if there was an iteration error
	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v", err)
		return nil, err
	}

	if len(messages) == 0 {
		log.Println("No messages found.")
	}

	log.Printf("Successfully retrieved chat history: %+v", messages)

	return messages, nil
}

// --- SQL DB Create Command ---
// CREATE DATABASE IF NOT EXISTS chatapp;
// USE chatapp;
// CREATE TABLE messages (
//     id INT AUTO_INCREMENT PRIMARY KEY,
//     sender VARCHAR(255) NOT NULL,
//     content TEXT NOT NULL,
//     timestamp DATETIME NOT NULL
// );
