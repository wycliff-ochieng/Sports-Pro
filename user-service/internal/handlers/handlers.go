package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/wycliff-ochieng/internal/service"
)

type UserHandler struct {
	l *log.Logger
	p service.Profile
}

type UpdatePofileReq struct {
	UserID    int
	Firstname string
	Lastname  string
	avatar    string
	Email     string
	//Createdat string
	Updatedat time.Time
}

func NewUserHandler(l *log.Logger, p service.Profile) *UserHandler {
	return &UserHandler{
		l: l,
		p: p,
	}
}

func (u *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {}

func (u *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	u.l.Println(">>>updating user profile")

	ctx := context.WithCancel(r.context(),)

	var user *UpdatePofileReq

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "ERROR:unmarshaling data", http.StatusInternalServerError)
	}

	if user.Firstname == "" || user.Lastname == "" {
		http.Error(w, "please enter a valid name", http.StatusBadRequest)
	}

	profile, err := u.p.UpdateUserProfile(ctx, user.UserID, user.Firstname, user.Lastname,user.avatar,user.Email, user.Updatedat)

	w.Header().Add("Content-Type", "Application/json")
	json.NewEncoder(w).Encode(&profile)
}
