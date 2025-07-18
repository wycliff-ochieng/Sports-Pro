package service

import (
	"context"
	"fmt"

	"github.com/wycliff-ochieng/internal/database"
)

type UserService struct {
	db database.DBInterface
}

type EventsData struct {
	UserID int    `json:"userid"`
	Email  string `json:"email"`
}

func NewUser(db database.DBInterface) *UserService {
	return &UserService{db}
}

func (u *UserService) CreateUserProfile(ctx context.Context, userID int, email string) error {

	var event EventsData

	query := `INSERT INTO Users(userid,email) VALUES($1,$2)`

	_, err := u.db.ExecContext(ctx, query, event.UserID, event.Email)
	if err != nil {
		return fmt.Errorf("something happened: %v", err)
	}

	return nil
}
