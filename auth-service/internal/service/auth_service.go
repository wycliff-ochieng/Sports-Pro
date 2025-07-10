package service

import (
	"context"
	"database/sql"
	"errors"
	"sports/authservice/internal/database"
	"sports/authservice/internal/handlers"
	"sports/authservice/internal/models"

	"github.com/google/uuid"
)

var (
	ErrEmailExists = errors.New("email already esists")
	ErrNotFound    = errors.New("email does not exists")
)

type AuthService struct {
	db database.DBInterface
}

func NewAuthService(db database.DBInterface) *AuthService {
	return &AuthService{db}
}

func (s *AuthService) Register(ctx context.Context, id uuid.UUID, firstname string, lastname string, email string, password string) (*models.UserResponse, error) {

	var exists bool

	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM Users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists //fmt.Errorf("email already exists")
	}
	//create user
	user, err := models.NewUser(id, firstname, lastname, email, password)
	if err != nil {
		return nil, err
	}

	//insert into db
	query := "INSERT INTO Users(userid,firstname,lastname,email,password,createdat,updatedat) VALUES($1,$2,$3,$4,$5,$6,$7)"

	_, err = s.db.ExecContext(ctx, query, user.ID, user.FirstName, user.LastName, user.Email, user.Password, user.CreatedAT, user.UpdatedAT)
	if err != nil {
		return nil, err
	}
	return &models.UserResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAT,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (*handlers.TokenPair, *models.UserResponse, error) {
	var user models.User

	query := `SELECT id, email,password,firstname,lastname,createdat,updatedat FROM Users WHERE email = $1`

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAT,
		&user.UpdatedAT,
	)

	if err == sql.ErrNoRows {
		return nil, nil, ErrNotFound
	}
	return nil, nil, nil
}
