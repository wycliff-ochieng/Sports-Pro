package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github/wycliff-ochieng/internal/models"
	"github/wycliff-ochieng/internal/service"
	"github/wycliff-ochieng/middleware"
	"log"
	"net/http"
	"time"

	auth "github.com/wycliff-ochieng/sports-common-package/middleware"

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

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "failed to get userID", http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		h.l.Println("ERROR while marshaling the data")
		http.Error(w, "Error Decoding request input", http.StatusInternalServerError)
		return
	}

	team, err := h.t.CreateTeam(ctx, userID, create.TeamID, create.Name, create.Sport, create.Description, create.Createdat, create.Updatedat)
	if err != nil {
		h.l.Printf("somethig is wrong in the service transaction: %s", err)
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
	userUUID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "UUID for the user is missing", http.StatusBadRequest)
		h.l.Printf("GET USER UUID FROM CONTEXT ERROR: %v", err)
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

	teamIDStr := vars["team_id"]

	teamID, err := uuid.Parse(teamIDStr)

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Error: userID validation failed", http.StatusExpectationFailed)
		h.l.Printf("ERROR: %s", err)
		return
	}
	log.Println(userID, teamID)

	team, err := h.t.GetTeamDetails(ctx, userID, teamID) //change team id back to UUID not string
	if err != nil {
		log.Printf("error due to: %s", err)
		http.Error(w, "Error: ", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&team)

}

// GET :: api/teamid/members -> get a list/roster of members of a certain team
func (h *TeamHandler) GetTeamsMembers(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Fetching all members for a give teams")
	//must be a member of that team
}

func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Updating team information")

	ctx := r.Context()
	vars := mux.Vars(r)

	teamIDStr := vars["team_id"] // parse this team ID to UUID

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		http.Error(w, "ERROR: failed parsing this to uuid", http.StatusExpectationFailed)
		return
	}

	userID, err := auth.GetUserUUIDFromContext(ctx)
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
		log.Printf("FAILING DUE TO: %v", err)
		http.Error(w, "update team service transaction error", http.StatusFailedDependency)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&team)
}

// POST :: add members to a team -> RBAC(coach /manager role)
func (h *TeamHandler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Adding Members to a Team initiated successfuly")

	vars := mux.Vars(r)

	ctx := r.Context()

	teamIDStr := vars["team_id"] //convert this teamID to UUId - > was to write some parse function to convert this

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		http.Error(w, "failed to chnage teamID to uuid", http.StatusFailedDependency)
		return
	}

	//userID from context

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		log.Fatalf("UserID not found in context:%v", err)
		return
	}
	/*role, err := middleware.GetUserRoleFromContext(ctx)
	if err != nil {
		http.Error(w, "Error : issue with roles for this user", http.StatusBadRequest)
		return
	}*/

	//validate if userID id a valid UUUID format ,role is a valid type
	if userID == uuid.Nil {
		log.Fatal("Empyt userID, Please input User UUID")
	}

	var addMemberReq models.AddMemberReq

	err = json.NewDecoder(r.Body).Decode(&addMemberReq)
	if err != nil {
		http.Error(w, "failed to decode request message", http.StatusExpectationFailed)
		return
	}

	//call service layer =
	addedMember, err := h.t.AddTeamMember(ctx, teamID, userID, addMemberReq)
	if err != nil {
		http.Error(w, "ERROR: something wrong with addTeamMember subscriptio", http.StatusExpectationFailed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&addedMember)

}

func (h *TeamHandler) GetTeamRoster(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Fetching all team members and their profiles")

	ctx := r.Context()

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "failed to get userID from context", http.StatusFailedDependency)
		return
	}

	vars := mux.Vars(r)

	teamIDStr := vars["team_id"]

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		h.l.Printf("invalid team_id path parameter: %v", err)
		http.Error(w, "invalid team id", http.StatusBadRequest)
		return
	}

	members, err := h.t.GetTeamMebers(ctx, teamID, userID)
	if err != nil {
		http.Error(w, "error performing getTeamMemebeers database transaction", http.StatusExpectationFailed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&members)
}

func (h *TeamHandler) UpdateTeamMember(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Update a team members role by either coach / manager ")

	ctx := r.Context()

	vars := mux.Vars(r)

	teamIDStr := vars["teamid"]
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		log.Fatal("failed to convert string to uuid for teamid")
	}

	//get req userId from req context

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "failed to get userID from context", http.StatusNotFound)
	}

	var updateMember models.UpdateTeamMemberReq

	err = json.NewDecoder(r.Body).Decode(&updateMember)
	if err != nil {
		http.Error(w, "decode team member roles daata", http.StatusExpectationFailed)
		return
	}

	//calll user service
	member, err := h.t.UpdateTeamMembersRoles(ctx, userID, teamID, updateMember)
	if err != nil {
		http.Error(w, "update datatabase operation failed", http.StatusFailedDependency)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&member)
}

func (h *TeamHandler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	h.l.Println("Deleteing team member handler")

	ctx := r.Context()

	vars := mux.Vars(r)

	teamIDStr := vars["teamid"]
	userIDStr := vars["userid"]

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		http.Error(w, "Error changing teamId to uuid", http.StatusExpectationFailed)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Error changing userId to uuid", http.StatusExpectationFailed)
		return
	}

	reqUserID, err := middleware.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "failed to get requester's userId from context", http.StatusNotFound)
	}

	//call user service
	_, err = h.t.RemoveMember(ctx, reqUserID, userID, teamID)
	if err != nil {
		h.l.Println("cannot delete member from this table")
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "user with this ID was not found", http.StatusNotFound)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
