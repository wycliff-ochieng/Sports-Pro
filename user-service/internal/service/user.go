package service

import (
	"github.com/wycliff-ochieng/internal/models"
)

type UserService struct{}

func NewUser() *UserService {
	return &UserService{}
}

func (u *UserService) CreateUserProfile(userID int, email string) (*models.Profile, error) {

	return &models.Profile{}, nil
}
