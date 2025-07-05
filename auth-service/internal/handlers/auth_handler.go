package handlers

import (
	"log"
	"net/http"
)

type AuthHandler struct {
	l *log.Logger
}

func NewAuthHandler(l *log.Logger) *AuthHandler {
	return &AuthHandler{l}
}

func (a *AuthHandler) Register(w http.ResponseWriter, r *http.Request) error {
	a.l.Println(">>>REGISTER USER HANDLE TOUCHED")
	return nil
}
