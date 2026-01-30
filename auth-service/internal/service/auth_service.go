package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	auth "sports/authservice/internal/auth"
	"sports/authservice/internal/database"
	"sports/authservice/internal/models"
	"time"

	"github.com/google/uuid"
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

func (s *AuthService) Register(ctx context.Context, firstname string, lastname string, email string, password string) (*models.UserResponse, error) {

	var exists bool

	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM Users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists //fmt.Errorf("email already exists")
	}
	//create user
	user, err := models.NewUser(0, firstname, lastname, email, password)
	if err != nil {
		return nil, err
	}

	//insert into db
	query := "INSERT INTO Users(firstname,lastname,email,password,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,userid"

	role_query := "INSERT INTO user_roles(user_id,role_id) VALUES($1,$2)"

	var newUserID int
	var newUserUUID uuid.UUID

	defaultRoleID := 1

	err = s.db.QueryRowContext(ctx, query, user.FirstName, user.LastName, user.Email, user.Password, user.CreatedAT, user.UpdatedAT).Scan(&newUserID, &newUserUUID)
	if err != nil {
		return nil, err
	}

	log.Printf("DEBUG: New user created with internal ID: %d and external UUID: %s", newUserID, newUserUUID)

	user.ID = newUserID
	user.UserID = newUserUUID

	log.Printf("DEBUG: Attempting to assign role ID %d to user with internal ID %d", defaultRoleID, newUserID)

	_, err = s.db.ExecContext(ctx, role_query, newUserUUID, defaultRoleID)
	if err != nil {
		return nil, err
	}

	return &models.UserResponse{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		CreatedAt: user.CreatedAT,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email string, password string) (*auth.TokenPair, *models.UserResponse, error) {
	var user models.User

	query := `SELECT id,userid, email,password,firstname,lastname,created_at,updated_at FROM Users WHERE email = $1`

	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.UserID,
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

	role, err := s.FetchUserRoles(ctx, user.UserID)
	if err != nil {
		return nil, nil, err
	}

	//generate token pair
	token, err := auth.GenerateTokenPair(
		user.ID,
		user.UserID,
		role,
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

func (s *AuthService) FetchUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `SELECT r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id=$1`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var roles []string

	for rows.Next() {
		var rolename string
		if err := rows.Scan(&rolename); err != nil {
			return nil, fmt.Errorf("something happened when fetching user roles,%v", err)
		}

		roles = append(roles, rolename)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through the rows:%v", err)
	}
	return roles, err
}
