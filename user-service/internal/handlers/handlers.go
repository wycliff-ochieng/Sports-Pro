package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	//"github.com/aws/aws-sdk-go-v2/aws/middleware/private/metrics/middleware"

	"github.com/wycliff-ochieng/internal/service"
	"github.com/wycliff-ochieng/middleware"
)

type UserHandler struct {
	l *log.Logger
	p service.Profile
}

func NewUserHandler(l *log.Logger, p service.Profile) *UserHandler {
	return &UserHandler{
		l: l,
		p: p,
	}
}

func (u *UserHandler) GetUserProfileByUUID(w http.ResponseWriter, r *http.Request) {

	u.l.Println(">>>Getting user profile by UUiD handler ")

	userUUID, err := middleware.GetUserUUIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "cant get user UUID from context", http.StatusExpectationFailed)
		return
	}

	//call service layer
	profile, err := u.p.GetUserProfileByUUID(r.Context(), userUUID)
	if err != nil {
		http.Error(w, "cant get profile from the database", http.StatusExpectationFailed)
		return
	}

	//respond with profile data
	w.Header().Add("Context-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&profile)

	return

}

func (u *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	u.l.Println(">>>updating user profile")

	//ctx := context.WithDeadline(pctx,deadline)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var user *service.UpdatePofileReq

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "ERROR:unmarshaling data", http.StatusInternalServerError)
	}

	if user.Firstname == "" || user.Lastname == "" {
		http.Error(w, "please enter a valid name", http.StatusBadRequest)
	}

	profile, err := u.p.UpdateUserProfile(ctx, user.UserID, user.Firstname, user.Lastname, user.Email, user.Updatedat)
	if err != nil {
		u.l.Printf("error from user service: %v", err)
		http.Error(w, "Error updating user profile", http.StatusExpectationFailed)
		return
	}

	w.Header().Add("Content-Type", "Application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&profile)
}
