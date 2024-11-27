package server

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var users = map[string]Login{}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid method", er)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username) < 1 || len(password) < 8 {
		er := http.StatusNotAcceptable
		http.Error(w, "Invalid username/password", er)
		return
	}

	if _, ok := users[username]; ok {
		er := http.StatusConflict
		http.Error(w, "User already exists", er)
		return
	}

	hashedPassword, _ := hashPassword(password)
	users[username] = Login{
		HashedPassword: hashedPassword,
	}

	log.Printf("User registered successfully")
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid request method", er)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, ok := users[username]
	if !ok || !checkPasswordHash(password, user.HashedPassword) {
		er := http.StatusUnauthorized
		http.Error(w, "Invalid username or password", er)
		return
	}

	sessionToken := generateToken(32)
	csrfToken := generateToken(32)

	// Sets the session cookies
	// This will be automatically sent by the browser for any requests to our endpoints. Hence this introduces CSRF vulnerabilities.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true, // Ensures the session token cant be accessed by front-end JavaScript and only sent during HTTP requests
	})

	// Sets the CSRF Token
	// Because of Same-Origin Policy malicious websites cannot access this and only we will be able to get and attach the csrf-token to the request header.
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false, // Needs to be accessable client side to be added to request headers
	})

	user.SessionToken = sessionToken
	user.CSRFToken = csrfToken
	users[username] = user

	log.Println("Login Successfull")
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
	if err := authorize(r); err != nil {
		er := http.StatusUnauthorized
		http.Error(w, "Unauthorized", er)
		return
	}

	// Clear Auth Cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
	})

	// Clear Auth Data from DB
	username := r.FormValue("username")
	user, _ := users[username]
	user.SessionToken = ""
	user.CSRFToken = ""
	users[username] = user

	fmt.Fprintln(w, "Logged out.")
}

func Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		er := http.StatusMethodNotAllowed
		http.Error(w, "Invalid requst method", er)
		return
	}

	if err := authorize(r); err != nil {
		er := http.StatusUnauthorized
		http.Error(w, "Unauthorized", er)
		return
	}

	username := r.FormValue("username")
	fmt.Fprintf(w, "Authorized, welcome %s", username)
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

func authorize(r *http.Request) error {
	var AuthError = errors.New("Unauthorized")

	username := r.FormValue("username")
	user, ok := users[username]
	if !ok {
		log.Println("User Invalid")
		return AuthError
	}

	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" || sessionToken.Value != user.SessionToken {
		log.Println("Session Token Invalid")
		return AuthError
	}

	csrfToken := r.Header.Get("X-CSRF-Token")
	if csrfToken != user.CSRFToken || csrfToken == "" {
		log.Println("CSRF Token Invalid")
		return AuthError
	}

	return nil
}
