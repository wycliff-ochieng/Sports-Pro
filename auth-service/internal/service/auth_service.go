package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	auth "sports/authservice/internal/auth"
	"sports/authservice/internal/database"
	"sports/authservice/internal/models"
	"time"
	//"github.com/google/uuid"
)

var (
	ErrEmailExists     = errors.New("email already esists")
	ErrNotFound        = errors.New("email does not exists")
	ErrInvalidPassword = errors.New("incorrect passowrd")
)

type AuthService struct {
	db database.DBInterface
}

func NewAuthService(db database.DBInterface) *AuthService {
	return &AuthService{
		db: db,
	}
}

func (s *AuthService) Register(ctx context.Context, id int, firstname string, lastname string, email string, password string) (*models.UserResponse, error) {

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

func (s *AuthService) Login(ctx context.Context, email string, password string) (*auth.TokenPair, *models.UserResponse, error) {
	var user models.User

	query := `SELECT userid, email,password,firstname,lastname,createdat,updatedat FROM Users WHERE email = $1`

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
	if err != nil {
		return nil, nil, fmt.Errorf("user not found:%v", err)
	}
	//compare password
	if err := user.ComparePassword(password); err != nil {
		return nil, nil, fmt.Errorf("passwords do no match: %v", err)
	}

	//generate token pair
	token, err := auth.GenerateTokenPair(
		user.ID,
		user.Email,
		string(auth.JWTSecret),     //jwtsecret
		string(auth.RefreshSecret), //refreshsecret
		time.Hour*24,
		time.Hour*24*7,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %v", err)
	}
	return token, &models.UserResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAT,
	}, nil
}
