package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go-chat-app/models"

	_ "github.com/go-sql-driver/mysql"
)

// DBInterface defines database operations.
// Defines an interface that represents the database operations available. This allows us to decouple the application logic from our specific database implementation making a db switch easier.
type DBInterface interface {
	SaveMessage(msg models.Message) error
	GetChatHistory() ([]models.Message, error)
	DeleteAllMessages() error
	SaveUser(username, hashedPassword string) error
	GetUserByUsername(username string) (models.User, error)
	UpdateSessionAndCSRF(userID int, sessionToken, csrfToken string) error
	ClearSession(userID int) error
	GetUserBySessionToken(sessionToken string) (models.User, error)
}

// MySQLDB implements DBInterface (by having the same methods) for a MySQL database.
// Called wrapper struct or database abstraction struct
// This encapsulate the database connection (*sql.DB) inside a struct, instead of relying on a global variable.
// Doing so ensures stateful management of the database connection.
type MySQLDB struct {
	db *sql.DB
}

// NewMySQLDB creates a new instance of MySQLDB with a live mysql database connection.
func NewMySQLDB(dsn string) (*MySQLDB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB connection: %w", err)
	}

	// Retry up to 10 times with 5s wait
	for i := 0; i < 10; i++ {
		if err = db.Ping(); err == nil {
			break
		}
		log.Printf("Failed to connect to database: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("could not connect to database after 10 attempts: %w", err)
	}

	return &MySQLDB{db: db}, nil
}

// SaveMessage saves a chat message to the database.
func (m *MySQLDB) SaveMessage(msg models.Message) error { // Method receiver used here. m is convention or db
	_, err := m.db.Exec(
		"INSERT INTO messages (sender, content, timestamp) VALUES (?, ?, ?)",
		msg.Sender, msg.Content, msg.Timestamp,
	)
	return err
}

// GetChatHistory retrieves chat history messages from the database.
func (m *MySQLDB) GetChatHistory() ([]models.Message, error) {
	log.Println("Attempting to get chat history from MySQL database.")
	rows, err := m.db.Query("SELECT sender, content, timestamp FROM messages ORDER BY timestamp ASC")
	if err != nil {
		log.Printf("SQL error: %v", err)
		return nil, err
	}
	defer rows.Close()

	log.Println("MySQL db queried.")

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
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

// DeleteAllMessages deletes all chat messages from the database
func (m *MySQLDB) DeleteAllMessages() error {
	_, err := m.db.Exec("DELETE FROM messages")
	if err != nil {
		return fmt.Errorf("failed to delete all messages: %w", err)
	}
	return nil
}

// SaveUser saves user and security information to the database
func (m *MySQLDB) SaveUser(username, hashedPassword string) error {
	_, err := m.db.Exec(
		"INSERT INTO users (username, hashed_password) VALUES (?, ?)",
		username, hashedPassword,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("username already exists: %w", err)
		}
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

// GetUserByUsername will get a user from a username
func (m *MySQLDB) GetUserByUsername(username string) (models.User, error) {
	var user models.User
	err := m.db.QueryRow(
		`SELECT id, username, hashed_password,
                COALESCE(session_token, '') AS session_token,
                COALESCE(csrf_token, '') AS csrf_token
         FROM users WHERE username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.HashedPassword, &user.SessionToken, &user.CSRFToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("user not found: %w", err)
		}
		return models.User{}, fmt.Errorf("failed to retrieve user: %w", err)
	}
	return user, nil
}

// UpdateSessionAndCSRF will update he sessions and csrf token information for a given user in the database
func (m *MySQLDB) UpdateSessionAndCSRF(userID int, sessionToken, csrfToken string) error {
	_, err := m.db.Exec(
		"UPDATE users SET session_token = ?, csrf_token = ? WHERE id = ?",
		sessionToken, csrfToken, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session and CSRF tokens for userID %d: %w", userID, err)
	}
	return nil
}

// ClearSession clears user auth and csrf token data from a user when that sessions ends. e.g when logging out
func (m *MySQLDB) ClearSession(userID int) error {
	_, err := m.db.Exec(
		"UPDATE users SET session_token = '', csrf_token = '' WHERE id = ?",
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear session for userID %d: %w", userID, err)
	}
	return nil
}

// Gets a user from their session token
func (m *MySQLDB) GetUserBySessionToken(sessionToken string) (models.User, error) {
	var user models.User
	err := m.db.QueryRow(
		"SELECT id, username, session_token, csrf_token FROM users WHERE session_token = ?",
		sessionToken,
	).Scan(&user.ID, &user.Username, &user.SessionToken, &user.CSRFToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("session token not found: %w", err)
		}
		return models.User{}, fmt.Errorf("failed to retrieve user by session token: %w", err)
	}
	return user, nil
}
