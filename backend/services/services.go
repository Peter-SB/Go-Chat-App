package services

import (
	"go-chat-app/auth"
	"go-chat-app/db"
)

type Services struct {
	DB   db.DBInterface
	Auth auth.AuthServiceInterface
}
