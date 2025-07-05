package service

import "sports/authservice/internal/database"

type AuthService struct {
	db database.DBInterface
}
