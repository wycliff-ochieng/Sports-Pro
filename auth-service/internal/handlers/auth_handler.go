package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"sports/authservice/internal/auth"
	"sports/authservice/internal/models"
	internal "sports/authservice/internal/producer"
	"sports/authservice/internal/service"

	"github.com/google/uuid"
	//"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, firstname, lastname, email, password string) (*models.UserResponse, error)
	Login(ctx context.Context, email, password string) (*auth.TokenPair, *models.UserResponse, error)
}

type AuthHandler struct {
	l  *log.Logger
	As AuthService
	p  internal.KafkaProducer
}

type RegisterReq struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthenticationResponse struct {
	User         interface{}
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type UserCreatedEvent struct {
	UserID    uuid.UUID `json:"userid"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
}

func NewAuthHandler(l *log.Logger, as AuthService, p internal.KafkaProducer) *AuthHandler {
	return &AuthHandler{
		l:  l,
		As: as,
		p:  p,
	}
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	a.l.Println(">>>REGISTER USER HANDLE TOUCHED")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var RegisterUser *RegisterReq

	err := json.NewDecoder(r.Body).Decode(&RegisterUser)
	if err != nil {
		a.l.Printf("error:%v", err)
		http.Error(w, "Error decoding data", http.StatusInternalServerError)
		return
	}

	//validate if info is in correct format
	if RegisterUser.FirstName == "" || RegisterUser.LastName == "" || RegisterUser.Email == "" || RegisterUser.Password == "" {
		http.Error(w, "Error:", http.StatusExpectationFailed)
		return
	}

	user, err := a.As.Register(ctx, RegisterUser.FirstName, RegisterUser.LastName, RegisterUser.Email, RegisterUser.Password)
	if err == service.ErrEmailExists {
		http.Error(w, "ERROR: email already exists", http.StatusExpectationFailed)
		return
	}
	if err != nil {
		a.l.Printf("failed to register user:%v", err)
		http.Error(w, "unable to register user", http.StatusBadRequest)
		return
	}

	//after successfull event creation ,create user created event

	event := UserCreatedEvent{
		UserID:    user.UserID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	err = a.p.PublishUserCreation(ctx, event)
	if err != nil {
		a.l.Println("CRITICAL Failed to publish usercreation event")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&user)
}

func (a *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error picking up json response", http.StatusInternalServerError)
		return
	}

	//validate user input
	if req.Email == " " || req.Password == " " {
		http.Error(w, "email or password required", http.StatusExpectationFailed)
		return
	}

	//authenticate user (user service transactions)
	//token, user, err :=

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	token, user, err := a.As.Login(ctx, req.Email, req.Password)
	if err == service.ErrNotFound || err == service.ErrInvalidPassword {
		http.Error(w, "USER NOT FOUND,INVALID PASSWORD", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "FAILED TO SIGN IN", http.StatusInternalServerError)
		a.l.Printf("reason: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthenticationResponse{
		User:         user,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}
