package auth_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-chat-app/auth"
	"go-chat-app/db"

	"golang.org/x/crypto/bcrypt"
)

// Tests for the auth service using the mock db

func setupAuthService() (*auth.AuthService, *db.MockDB) {
	mockDB := db.NewMockDB()
	return auth.NewAuthService(mockDB), mockDB
}

func TestRegister_Success(t *testing.T) {
	service, _ := setupAuthService()

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("username=user1&password=securepassword"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	service.Register(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestRegister_InvalidInput(t *testing.T) {
	service, _ := setupAuthService()

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("username=user1&password=123"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	service.Register(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotAcceptable {
		t.Errorf("expected status %d, got %d", http.StatusNotAcceptable, resp.StatusCode)
	}
}

func TestRegister_UsernameConflict(t *testing.T) {
	service, mockDB := setupAuthService()
	mockDB.SaveUser("user1", "hashedpassword")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("username=user1&password=securepassword"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	service.Register(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, resp.StatusCode)
	}
}

func TestLoginUser_Success(t *testing.T) {
	service, mockDB := setupAuthService()

	password := "securepassword"
	hashedPasswordBytes, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
	hashedPassword := string(hashedPasswordBytes)
	mockDB.SaveUser("user1", hashedPassword)

	mockDB.UpdateSessionAndCSRF(1, "session123", "csrf123")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("username=user1&password="+password))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	service.LoginUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	cookies := resp.Cookies()
	if len(cookies) != 2 {
		t.Errorf("expected 2 cookies, got %d", len(cookies))
	}
}

func TestLoginUser_InvalidCredentials(t *testing.T) {
	service, _ := setupAuthService()

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("username=user1&password=wrongpassword"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	service.LoginUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestLogoutUser_Success(t *testing.T) {
	service, mockDB := setupAuthService()
	mockDB.SaveUser("user1", "hashedpassword")
	mockDB.UpdateSessionAndCSRF(1, "session123", "csrf123")

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "session123"})
	req.Header.Set("X-CSRF-Token", "csrf123")
	w := httptest.NewRecorder()

	service.LogoutUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Value != "" {
			t.Errorf("expected cookie value to be cleared, got %s", cookie.Value)
		}
	}
}

func TestLogoutUser_Unauthorised(t *testing.T) {
	service, _ := setupAuthService()

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	service.LogoutUser(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestProfile_Success(t *testing.T) {
	service, mockDB := setupAuthService()
	mockDB.SaveUser("user1", "hashedpassword")
	mockDB.UpdateSessionAndCSRF(1, "session123", "csrf123")

	req := httptest.NewRequest(http.MethodPost, "/profile", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "session123"})
	req.Header.Set("X-CSRF-Token", "csrf123")
	w := httptest.NewRecorder()

	service.Profile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestProfile_Unauthorised(t *testing.T) {
	service, _ := setupAuthService()

	req := httptest.NewRequest(http.MethodPost, "/profile", nil)
	w := httptest.NewRecorder()

	service.Profile(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestSessionCheck_Success(t *testing.T) {
	service, mockDB := setupAuthService()

	mockDB.SaveUser("user1", "hashedpassword")
	mockDB.UpdateSessionAndCSRF(1, "valid-session-token", "valid-csrf-token")

	req := httptest.NewRequest(http.MethodGet, "/session-check", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-session-token"})

	w := httptest.NewRecorder()

	service.SessionCheck(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	expectedBody := `{"username": "user1"}`
	body := w.Body.String()
	if body != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, body)
	}
}

func TestSessionCheck_InvalidSessionToken(t *testing.T) {
	service, mockDB := setupAuthService()

	mockDB.SaveUser("user1", "hashedpassword")

	req := httptest.NewRequest(http.MethodGet, "/session-check", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid-session-token"})

	w := httptest.NewRecorder()

	service.SessionCheck(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}
