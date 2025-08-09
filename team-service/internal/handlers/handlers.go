package handlers

import (
	"context"
	"encoding/json"
	"github/wycliff-ochieng/internal/service"
	"log"
	"net/http"
	"time"
)

type TeamHandler struct {
	l *log.Logger
	t *service.TeamService
}

type createTeamReq struct {
	Name        string
	Sport       string
	Description string
}

func NewTeamHandler(l *log.Logger, t *service.TeamService) *TeamHandler {
	return &TeamHandler{
		l: l,
		t: t,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {

	h.l.Println("CREATE TEAM HANDLER NOW TOUCHED")

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var create *createTeamReq

	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		h.l.Println("ERROR while marshaling the data")
		http.Error(w, "Error Decoding request input", http.StatusInternalServerError)
		return
	}

	team, err := h.t.CreateTeam(ctx, create.Name, create.Sport, create.Description)
	if err != nil {
		h.l.Println("somethig is wrong in the service transaction")
		http.Error(w, "FAILED:Error creating team service", http.StatusFailedDependency)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&team)
}
