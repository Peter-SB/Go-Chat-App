package routes

import (
	"net/http"

	"go-chat-app/handlers"
	"go-chat-app/middleware"
	"go-chat-app/services"
)

func SetupRoutes(services *services.Services) {
	// Define allowed origins for use by cors middleware
	allowedOrigins := []string{
		"http://localhost:3000",
	}

	corsMiddleware := middleware.CORSMiddleware(allowedOrigins)

	http.Handle("/history", corsMiddleware(http.HandlerFunc(handlers.GetChatHistoryHandler(services))))
	http.Handle("/ws", corsMiddleware(http.HandlerFunc(handlers.HandleConnections(services))))

	http.HandleFunc("/register", services.Auth.Register)
	http.HandleFunc("/login", services.Auth.LoginUser)
	http.HandleFunc("/logout", services.Auth.LogoutUser)
	http.HandleFunc("/profile", services.Auth.Profile)
}
