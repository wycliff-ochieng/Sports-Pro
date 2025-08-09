package handlers

import (
	"github/wycliff-ochieng/internal/service"
	"log"
	"net/http"
)

type TeamHandler struct {
	l *log.Logger
	t *service.TeamService
}

func NewTeamHandler(l *log.Logger, t *service.TeamService) *TeamHandler {
	return &TeamHandler{
		l: l,
		t: t,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	return
}
