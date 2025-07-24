package service

import (
	"context"
	//"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/models"
)

type Profile interface {
	GetProfileByID(int) (*models.Profile, error)
	UpdateUserProfile(ctx context.Context, userID int, firstname, lastname, email, avatar string, updatedat time.Time) (*models.Profile, error)
}

type UserService struct {
	l  *log.Logger
	db database.DBInterface
	//dbTx *sql.DB
}

type EventsData struct {
	UserID int    `json:"userid"`
	Email  string `json:"email"`
}

func NewUserService(l *log.Logger, db database.DBInterface) *UserService {
	return &UserService{
		db: db,
		l:  l,
		//dbTx:dbTx,
	}
}

var (
	ErrNotFound     = errors.New("User Id not found")
	ErrUpdateFailed = errors.New("Failed to update user profile")
)

func (u *UserService) CreateUserProfile(ctx context.Context, userID int, email string) error {

	var event EventsData

	query := `INSERT INTO Users(userid,email) VALUES($1,$2)`

	_, err := u.db.ExecContext(ctx, query, event.UserID, event.Email)
	if err != nil {
		return fmt.Errorf("something happened: %v", err)
	}

	return nil
}

func (u *UserService) GetProfileByID(ctx context.Context, userID int) error {
	var event EventsData

	query := "SELECT * from profiles WHERE id=$1"

	_, err := u.db.ExecContext(ctx, query, event.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserService) UpdateUserProfile(ctx context.Context, userID int, firstname string, lastname string, email string, avatar string) (*models.Profile, error) {

	u.l.Printf("UPDATING USER PROFILE UNDERWAY FOR:%v", userID)

	//var exists bool

	//err := u.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM profiles WHERE EMAIL=$1)").Scan(&exists)
	//if err != nil {
	//	return nil, fmt.Errorf("errorr checking if email exists:%v", err)
	//}

	// Begin transaction
	transaction, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		u.l.Printf("Transaction undergone some errors:%v", err)
	}

	//defer rollback - if anything below panicks everything will be roollback
	defer transaction.Rollback()

	//check if user exists, call get user by ID

	//request data ID
	return nil, nil
}
