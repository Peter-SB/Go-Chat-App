package routes

import (
	"net/http"

	"go-chat-app/handlers"
	"go-chat-app/middleware"
	"go-chat-app/services"
)

func SetupRoutes(services *services.Services) {
	corsMiddleware := middleware.CORSMiddleware()

	http.Handle("/history", corsMiddleware(http.HandlerFunc(handlers.ChatHistoryHandler(services))))
	http.Handle("/ws", corsMiddleware(http.HandlerFunc(handlers.HandleConnections(services))))

	http.Handle("/register", corsMiddleware(http.HandlerFunc(services.Auth.Register)))
	http.Handle("/login", corsMiddleware(http.HandlerFunc(services.Auth.LoginUser)))
	http.Handle("/logout", corsMiddleware(http.HandlerFunc(services.Auth.LogoutUser)))
	http.Handle("/profile", corsMiddleware(http.HandlerFunc(services.Auth.Profile))) // Not used by frontend, just for test/demonstration purposes
}
