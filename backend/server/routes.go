package server

import (
	"go-chat-app/middleware"

	"net/http"
)

func SetupRoutes() {
	// Define allowed origins
	allowedOrigins := []string{
		"http://localhost:3000",
	}

	corsMiddleware := middleware.CORSMiddleware(allowedOrigins)

	http.Handle("/history", corsMiddleware(http.HandlerFunc(GetChatHistoryHandler)))
	http.Handle("/ws", corsMiddleware(http.HandlerFunc(HandleConnections)))

	http.HandleFunc("/register", Register)
	http.HandleFunc("/login", LoginUser)
	http.HandleFunc("/logout", LogoutUser)
	http.HandleFunc("/profile", Profile)
}
