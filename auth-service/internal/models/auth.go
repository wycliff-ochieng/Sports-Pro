package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int       `json:"id"`
	UserID      uuid.UUID `json:"userid"`
	FirstName   string    `json:"firstname"`
	LastName    string    `json:"lastname"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	PhoneNumber int32     `json:"phonenumber"`
	Metadata    Metadata  `json:"metadata"`
	CreatedAT   time.Time `json:"crreatedat"`
	UpdatedAT   time.Time `json:"updatedat"`
}

type Metadata struct {
	Location string `json:"location"`
}

type UserResponse struct {
	ID        int       `json:"id"`
	UserID    uuid.UUID `json:"userId"`
	FirstName string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdat"`
}

func NewUser(id int, firstname string, lastname string, email string, password string) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("HASH_ERROR:Failed to harsh password:%v", err)
	}
	return &User{
		ID:        id,
		FirstName: firstname,
		LastName:  lastname,
		Email:     email,
		Password:  string(hashedPassword),
	}, nil
}

func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
