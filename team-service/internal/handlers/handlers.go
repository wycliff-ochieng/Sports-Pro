package handlers

import (
	"context"
	"encoding/json"
	"github/wycliff-ochieng/internal/service"
	"github/wycliff-ochieng/middleware"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type TeamHandler struct {
	l *log.Logger
	t *service.TeamService
}

type createTeamReq struct {
	TeamID      uuid.UUID
	Name        string
	Sport       string
	Description string
	Createdat   time.Time
	Updatedat   time.Time
}

func NewTeamHandler(l *log.Logger, t *service.TeamService) *TeamHandler {
	return &TeamHandler{
		l: l,
		t: t,
	}
}

// POST :: api/teams/
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

	team, err := h.t.CreateTeam(ctx, create.TeamID, create.Name, create.Sport, create.Description, create.Createdat, create.Updatedat)
	if err != nil {
		h.l.Println("somethig is wrong in the service transaction")
		http.Error(w, "FAILED:Error creating team service", http.StatusExpectationFailed)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&team)
}

// GET :: api/teams/id - all teams for a single user
func (h *TeamHandler) GetTeams(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Getting all the teams for the logged in user: ")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	//teams , err := h.t.GetMyTeams()
	/*userID,  err := middleware.GetUserIDFromContext(ctx)
	if err != nil{
		http.Error(w,"No ID for this User, not allowed", http.StatusNotFound)
		return
	}*/
	userUUID, err := middleware.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "UUID for the user is missing", http.StatusBadRequest)
		return
	}

	myTeams, err := h.t.GetMyTeams(ctx, userUUID)
	if err != nil {
		http.Error(w, "Error while fetching teams for this user", http.StatusNotFound)
		return
	}

	//marshalling the data from database -> change to json
	type MyTeams struct {
		TeamID      uuid.UUID
		Name        string
		Sport       string
		Description string
		CreatedAt   time.Time
		JoinedAT    time.Time
		Role        string
	}
	// TODO :: finish this tomorrow
	json.NewEncoder(w).Encode(&myTeams)
}

// GET :: api/teams/{teamID} - get a detailed public profile for a single team
func (h *TeamHandler) GetTeamsByID(w http.ResponseWriter, r *http.Request) {
	return
}
