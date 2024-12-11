package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"go-chat-app/db"
	"go-chat-app/models"

	"golang.org/x/crypto/bcrypt"
)

// AuthServiceInterface defines the methods for the authentication service.
type AuthServiceInterface interface {
	Register(w http.ResponseWriter, r *http.Request)
	LoginUser(w http.ResponseWriter, r *http.Request)
	LogoutUser(w http.ResponseWriter, r *http.Request)
	Profile(w http.ResponseWriter, r *http.Request)
	Authorise(r *http.Request) (*models.User, error)
}

type AuthService struct {
	db db.DBInterface
}

func NewAuthService(db db.DBInterface) *AuthService {
	return &AuthService{db: db}
}

func (a *AuthService) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	log.Printf("Registering username: %s", username)

	if len(username) < 1 || len(password) < 4 {
		log.Printf("Invalid registration details - username: '%s', password length: %d", username, len(password))
		http.Error(w, "Invalid username or password (password must be at least 4 characters)", http.StatusNotAcceptable)
		return
	}

	// Check if the user already exists
	if _, err := a.db.GetUserByUsername(username); err == nil {
		log.Printf("Registration failed: username '%s' already exists", username)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Printf("Failed to hash password for user '%s': %v", username, err)
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	log.Println("Saving user...")

	// Save the user to the database
	err = a.db.SaveUser(username, hashedPassword)
	if err != nil {
		log.Printf("Error saving user '%s' to the database: %v", username, err)
		http.Error(w, "Error saving user", http.StatusInternalServerError)
		return
	}

	log.Println("User registered successfully")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully"))
}

func (a *AuthService) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("LoginUser error: invalid request method %s", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	log.Printf("Logging in username: %s", username)

	if username == "" || password == "" {
		log.Printf("LoginUser error: missing username or password. Username: %s", username)
		http.Error(w, "Missing username or password", http.StatusBadRequest)
		return
	}

	// Fetch user from database
	user, err := a.db.GetUserByUsername(username)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			log.Printf("Login failed: User not found with username '%s'", username)
		} else {
			http.Error(w, "Error retrieving user", http.StatusInternalServerError)
			log.Printf("Error retrieving user from database: %v", err)
		}
		return
	}

	// Validate password
	if !checkPasswordHash(password, user.HashedPassword) {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		log.Printf("Login failed: Invalid password for username '%s'", username)
		return
	}

	// Generate session and CSRF tokens
	sessionToken := generateToken(32)
	csrfToken := generateToken(32)

	// Sets the session cookies.
	// This will be automatically sent by the browser for any requests to our endpoints on the same domain.
	// Hence this introduces CSRF vulnerabilities because the cookie will automatically be sent allowing forged cross-origin requests.
	// HttpOnly and Secure flags mitigate risks like XSS and data interception.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,                    // Ensures the session token cant be accessed by front-end JavaScript and only sent during HTTP requests. Reducing XSS risk.
		Secure:   true,                    // Ensures that the cookie is only sent over HTTPS connections, preventing interception over insecure HTTP. If Secure is not set explicitly, the cookie will be sent over both HTTP and HTTPS.
		SameSite: http.SameSiteStrictMode, // Controls whether cookies are sent with cross-site requests, mitigating CSRF risks. The default for SameSite is unset, which allows cookies to be sent with cross-origin requests.
	})

	// Sets the CSRF Token
	// When the CSRF token is sent back to the server for authentication, the user must explisitly send it in a custom request header.
	// Because the custom request header (tippicaly called "X-CSRF-Token") is added by the client and not sent automaticaly, Same-Origin
	// Policy stops malicious websites from accessing this and only we are able to get and attach the csrf-token to the x-csrf-token request header.
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false, // Needs to be accessable client side to be added to request headers
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	// Update the user's session and CSRF tokens in the database
	err = a.db.UpdateSessionAndCSRF(user.ID, sessionToken, csrfToken)
	if err != nil {
		http.Error(w, "Error updating session", http.StatusInternalServerError)
		log.Printf("Error updating session: %v", err)
		return
	}

	log.Println("Login Successfull")
	w.WriteHeader(http.StatusOK)
}

func (a *AuthService) LogoutUser(w http.ResponseWriter, r *http.Request) {
	user, err := a.Authorise(r)
	if err != nil {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		return
	}

	// Clear Token Cookies
	a.setCookie(w, "session_token", "", true, true)
	a.setCookie(w, "csrf_token", "", false, true)

	// Clear session and CSRF tokens in the database
	err = a.db.ClearSession(user.ID)
	if err != nil {
		http.Error(w, "Error clearing session", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Logged out.")
}

func (a *AuthService) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	user, err := a.Authorise(r)
	if err != nil {
		http.Error(w, "Unauthorised", http.StatusUnauthorized)
		log.Printf("Error authorizing session: %v", err)
		return
	}

	fmt.Fprintf(w, "Authorised, welcome %s", user.Username)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10) // Cost = 10 means the password is hashed 2^10 times.
	// This is to slow down any attempt to "hash crack", ie, reverse engineer the password by making guesses and seeing if that matches the hashed password
	// Note: bcrypt also automaticaly handles salting to protect against prcomputed hash table attacks.

	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateToken(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes)
}

func (a *AuthService) Authorise(r *http.Request) (*models.User, error) {
	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" {
		log.Printf("Authorization failed: Missing or empty session token. Error: %v", err)
		return nil, errors.New("missing session token")
	}

	csrfToken := r.Header.Get("X-CSRF-Token")
	// If not present in the header, check the query parameter
	if csrfToken == "" {
		// Parse the query parameters
		queryParams, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			log.Printf("Invalid query parameters")
			return nil, errors.New("invalid query parameters")
		}
		csrfToken = queryParams.Get("csrf_token")
	}

	if csrfToken == "" {
		log.Println("Authorization failed: Missing CSRF token in request header.")
		return nil, errors.New("missing CSRF token")
	}

	// Use the session token to identify the user.
	user, err := a.db.GetUserBySessionToken(sessionToken.Value)
	if err != nil {
		log.Printf("Authorization failed: Unable to fetch user for session token %s. Error: %v", sessionToken.Value, err)
		return nil, errors.New("unauthorised")
	}

	if user.CSRFToken != csrfToken {
		log.Printf("Authorization failed: CSRF token mismatch for user %s. Expected: %s, Received: %s",
			user.Username, user.CSRFToken, csrfToken)
		return nil, errors.New("unauthorised")
	}

	log.Printf("Authorization successful for user: %s", user.Username)
	return &user, nil
}

func (a *AuthService) setCookie(w http.ResponseWriter, name, value string, httpOnly, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}
