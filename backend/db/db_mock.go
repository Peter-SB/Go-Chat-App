package db

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go-chat-app/models"
)

type MockDB struct {
	mu       sync.Mutex
	messages []models.Message
	users    map[string]models.User // keyed by username
	nextID   int
}

func NewMockDB() *MockDB {
	return &MockDB{
		messages: []models.Message{},
		users:    make(map[string]models.User),
		nextID:   1,
	}
}

// SaveMessage (mock) stores a chat message in memory.
func (m *MockDB) SaveMessage(msg models.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Timestamp handling if needed
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	m.messages = append(m.messages, msg)
	return nil
}

// GetChatHistory (mock) retrieves all stored messages.
func (m *MockDB) GetChatHistory() ([]models.Message, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy to avoid external modification
	history := make([]models.Message, len(m.messages))
	copy(history, m.messages)
	return history, nil
}

// DeleteAllMessages (mock) clears all messages.
func (m *MockDB) DeleteAllMessages() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = []models.Message{}
	return nil
}

// SaveUser (mock) saves a new user if it does not already exist.
func (m *MockDB) SaveUser(username, hashedPassword string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for existing user
	if _, exists := m.users[username]; exists {
		return fmt.Errorf("username already exists")
	}

	user := models.User{
		ID:             m.nextID,
		Username:       username,
		HashedPassword: hashedPassword,
		SessionToken:   "",
		CSRFToken:      "",
	}
	m.users[username] = user
	m.nextID++
	return nil
}

// GetUserByUsername (mock) retrieves a user by username.
func (m *MockDB) GetUserByUsername(username string) (models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[username]
	if !exists {
		return models.User{}, errors.New("user not found")
	}
	return user, nil
}

// UpdateSessionAndCSRF (mock) updates the session and CSRF token for a given user.
func (m *MockDB) UpdateSessionAndCSRF(userID int, sessionToken, csrfToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find user by ID
	var usernameToUpdate string
	for uname, user := range m.users {
		if user.ID == userID {
			usernameToUpdate = uname
			break
		}
	}

	if usernameToUpdate == "" {
		return errors.New("user not found")
	}

	user := m.users[usernameToUpdate]
	user.SessionToken = sessionToken
	user.CSRFToken = csrfToken
	m.users[usernameToUpdate] = user

	return nil
}

// ClearSession (mock) clears the session and csrf tokens from a user.
func (m *MockDB) ClearSession(userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find user by ID
	var usernameToClear string
	for uname, user := range m.users {
		if user.ID == userID {
			usernameToClear = uname
			break
		}
	}

	if usernameToClear == "" {
		return errors.New("user not found")
	}

	user := m.users[usernameToClear]
	user.SessionToken = ""
	user.CSRFToken = ""
	m.users[usernameToClear] = user

	return nil
}

// GetUserBySessionToken (mock) retrieves a user by their session token.
func (m *MockDB) GetUserBySessionToken(sessionToken string) (models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, user := range m.users {
		if strings.TrimSpace(user.SessionToken) == strings.TrimSpace(sessionToken) && sessionToken != "" {
			return user, nil
		}
	}

	return models.User{}, errors.New("session token not found")
}
