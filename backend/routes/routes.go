package routes

import (
	"net/http"

	"go-chat-app/auth"
	"go-chat-app/handlers"
	"go-chat-app/middleware"
)

func SetupRoutes() {
	// Define allowed origins
	allowedOrigins := []string{
		"http://localhost:3000",
	}

	corsMiddleware := middleware.CORSMiddleware(allowedOrigins)

	http.Handle("/history", corsMiddleware(http.HandlerFunc(handlers.GetChatHistoryHandler)))
	http.Handle("/ws", corsMiddleware(http.HandlerFunc(handlers.HandleConnections)))

	http.HandleFunc("/register", auth.Register)
	http.HandleFunc("/login", auth.LoginUser)
	http.HandleFunc("/logout", auth.LogoutUser)
	http.HandleFunc("/profile", auth.Profile)
}
