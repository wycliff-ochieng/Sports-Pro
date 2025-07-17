package handlers

import (
	"log"
	"net/http"
)

type UserHandler struct {
	l *log.Logger
}

func NewUserHandler(l *log.Logger) *UserHandler {
	return &UserHandler{l}
}

func (u *UserHandler) UpdateUserProfile(w http.ResponseWriter, r *http.Request) {}
