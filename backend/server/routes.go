package server

import "net/http"

func SetupRoutes() {
	http.HandleFunc("/history", GetChatHistoryHandler)

	http.HandleFunc("/ws", HandleConnections)

	http.HandleFunc("/register", Register)
	http.HandleFunc("/login", LoginUser)
	http.HandleFunc("/logout", LogoutUser)
	http.HandleFunc("/profile", Profile)
}
