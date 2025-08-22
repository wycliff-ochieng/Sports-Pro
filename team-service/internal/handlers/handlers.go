package handlers

import (
	"context"
	"encoding/json"
	"github/wycliff-ochieng/internal/models"
	"github/wycliff-ochieng/internal/service"
	"github/wycliff-ochieng/middleware"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TeamHandler struct {
	l *log.Logger
	t *service.TeamService
}

type createTeamReq struct {
	TeamID      uuid.UUID `json:"teamid"`
	Name        string    `json:"name"`
	Sport       string    `json:"sport"`
	Description string    `json:"description"`
	Createdat   time.Time `json:"createdat"`
	Updatedat   time.Time `json:"updatedat"`
}

type updateTeamDetailsReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&myTeams)
}

// GET :: api/teams/{teamID} - get a detailed public profile for a single team
func (h *TeamHandler) GetTeamsByID(w http.ResponseWriter, r *http.Request) {
	h.l.Println("info: getting team details for this specific team id")

	ctx := r.Context()

	vars := mux.Vars(r)

	teamID := vars["team_id"]

	userID, err := middleware.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Error: userID validation failed", http.StatusExpectationFailed)
		return
	}

	//	roles, err := middleware.GetUserRoleFromContext(ctx)
	//	if err != nil {
	//		http.Error(w, "Error: no role found for this user", http.StatusNotFound)
	//		return
	//	}

	/*	isMember,err := h.t.IsTeamMember(ctx,userID,teamID)
		if err != nil{
			return false,err
		}
		if !isMember{
			return nil,http.StatusNotFound
		}*/
	team, err := h.t.GetTeamByID(ctx, teamID, userID) //change team id back to UUID not string
	if err != nil {
		http.Error(w, "Error: ", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&team)

	/*
		if roles == "COACH" || roles == "MANAGER" {
			team, err := h.t.GetTeamByID(ctx, teamID, userID)
			if err != nil {
				http.Error(w, "Error: ", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&team)
		} else {

			log.Fatal("Not allowed to view all the teams")
		}
	*/
}

func (h *TeamHandler) GetTeamsMembers(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Fetching all members for a give teams")
	//must be a member of that team
}

func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Updating team information")

	ctx := r.Context()
	vars := mux.Vars(r)

	teamID := vars["teamID"] // parse this team ID to UUID

	userID, err := middleware.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Error fetching UUID from context", http.StatusNotFound)
		return
	}

	var update models.UpdateTeamReq
	//get userID from context
	//get teamID from url

	err = json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		log.Fatalf("Error decoding %v", err)
	}

	team, err := h.t.UpdateTeamDetails(ctx, teamID, userID, update)
	if err != nil {
		http.Error(w, "update team service transaction error", http.StatusFailedDependency)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&team)
}
