package service

import (
	"context"
	"fmt"

	"github.com/wycliff-ochieng/internal/database"
)

type Profile interface{}

type UserService struct {
	db database.DBInterface
}

type EventsData struct {
	UserID int    `json:"userid"`
	Email  string `json:"email"`
}

func NewUserService(db database.DBInterface) *UserService {
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

func (u *UserService) GetProfileByID(ctx context.Context,userID int) error{
	var event EventsData
	
	query := "SELECT * from profiles WHERE id=$1"

	_, err := u.db.ExecContext(ctx, query,event.UserID)
	if err != nil{
		return err
	}

	return nil
}
