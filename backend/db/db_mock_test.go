package db_test

import (
	"testing"
	"time"

	"go-chat-app/db"
	"go-chat-app/models"
)

func TestSaveMessage(t *testing.T) {
	mockDB := db.NewMockDB()
	msg := models.Message{
		Sender:    "user1",
		Content:   "Hello, World!",
		Timestamp: time.Now(),
	}

	err := mockDB.SaveMessage(msg)
	if err != nil {
		t.Fatalf("SaveMessage failed: %v", err)
	}

	history, _ := mockDB.GetChatHistory()
	if len(history) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(history))
	}
	if history[0].Content != "Hello, World!" {
		t.Errorf("Expected message content 'Hello, World!', got '%s'", history[0].Content)
	}
}

func TestGetChatHistory(t *testing.T) {
	mockDB := db.NewMockDB()
	msg1 := models.Message{Sender: "user1", Content: "Hi!", Timestamp: time.Now()}
	msg2 := models.Message{Sender: "user2", Content: "Hello!", Timestamp: time.Now()}

	mockDB.SaveMessage(msg1)
	mockDB.SaveMessage(msg2)

	history, err := mockDB.GetChatHistory()
	if err != nil {
		t.Fatalf("GetChatHistory failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(history))
	}
}

func TestDeleteAllMessages(t *testing.T) {
	mockDB := db.NewMockDB()

	// Add some messages
	mockDB.SaveMessage(models.Message{Sender: "user1", Content: "Hello!", Timestamp: time.Now()})
	mockDB.SaveMessage(models.Message{Sender: "user2", Content: "Hi there!", Timestamp: time.Now()})

	// Verify messages were added
	history, err := mockDB.GetChatHistory()
	if err != nil {
		t.Fatalf("GetChatHistory failed: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(history))
	}

	// Delete all messages
	err = mockDB.DeleteAllMessages()
	if err != nil {
		t.Fatalf("DeleteAllMessages failed: %v", err)
	}

	// Verify all messages were deleted
	history, err = mockDB.GetChatHistory()
	if err != nil {
		t.Fatalf("GetChatHistory failed after deletion: %v", err)
	}
	if len(history) != 0 {
		t.Fatalf("Expected 0 messages after deletion, got %d", len(history))
	}
}

func TestSaveUser(t *testing.T) {
	mockDB := db.NewMockDB()

	err := mockDB.SaveUser("user1", "hashedpassword123")
	if err != nil {
		t.Fatalf("SaveUser failed: %v", err)
	}

	err = mockDB.SaveUser("user1", "anotherpassword")
	if err == nil {
		t.Fatal("Expected error for duplicate username, got nil")
	}
}

func TestGetUserByUsername(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SaveUser("user1", "hashedpassword123")

	user, err := mockDB.GetUserByUsername("user1")
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if user.Username != "user1" {
		t.Errorf("Expected username 'user1', got '%s'", user.Username)
	}

	_, err = mockDB.GetUserByUsername("nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent user, got nil")
	}
}

func TestUpdateSessionAndCSRF(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SaveUser("user1", "hashedpassword123")
	user, _ := mockDB.GetUserByUsername("user1")

	err := mockDB.UpdateSessionAndCSRF(user.ID, "session123", "csrf123")
	if err != nil {
		t.Fatalf("UpdateSessionAndCSRF failed: %v", err)
	}

	updatedUser, _ := mockDB.GetUserByUsername("user1")
	if updatedUser.SessionToken != "session123" || updatedUser.CSRFToken != "csrf123" {
		t.Error("Session and CSRF tokens were not updated correctly")
	}
}

func TestClearSession(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SaveUser("user1", "hashedpassword123")
	user, _ := mockDB.GetUserByUsername("user1")

	mockDB.UpdateSessionAndCSRF(user.ID, "session123", "csrf123")
	mockDB.ClearSession(user.ID)

	updatedUser, _ := mockDB.GetUserByUsername("user1")
	if updatedUser.SessionToken != "" || updatedUser.CSRFToken != "" {
		t.Error("Session and CSRF tokens were not cleared correctly")
	}
}

func TestGetUserBySessionToken(t *testing.T) {
	mockDB := db.NewMockDB()
	mockDB.SaveUser("user1", "hashedpassword123")
	user, _ := mockDB.GetUserByUsername("user1")

	mockDB.UpdateSessionAndCSRF(user.ID, "session123", "csrf123")
	retrievedUser, err := mockDB.GetUserBySessionToken("session123")
	if err != nil {
		t.Fatalf("GetUserBySessionToken failed: %v", err)
	}
	if retrievedUser.Username != "user1" {
		t.Errorf("Expected username 'user1', got '%s'", retrievedUser.Username)
	}

	_, err = mockDB.GetUserBySessionToken("invalidsession")
	if err == nil {
		t.Fatal("Expected error for invalid session token, got nil")
	}
}
