package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/internal/service"
	auth "github.com/wycliff-ochieng/sports-proto/middleware"
)

type EventHandler struct {
	logger *log.Logger
	es     *service.EventService
}

func NewEventHandler(l *log.Logger, es *service.EventService) *EventHandler {
	return &EventHandler{
		logger: l,
		es:     es,
	}
}

// POST /api/event/add
func (eh *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	eh.logger.Println("create event handle now firing up....")

	ctx := r.Context()

	var createReq models.CreateEventReq

	err := json.NewDecoder(r.Body).Decode(&createReq)
	if err != nil {
		http.Error(w, "Error decoding create event body data", http.StatusInternalServerError)
		return
	}

	reqUserID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "failed to get the requester userID from context", http.StatusExpectationFailed)
		return
	}

	if createReq.TeamID == uuid.Nil {
		http.Error(w, "empty teamID , please provide one", http.StatusExpectationFailed)
		return
	}
	if createReq.EventType == "" {
		http.Error(w, "please input event type", http.StatusExpectationFailed)
		return
	}

	event, err := eh.es.CreateTeamEvent(ctx, reqUserID, createReq.EventID, createReq.TeamID, createReq.EventType, createReq.StartTime, createReq.EndTime)
	if err != nil {
		http.Error(w, "Issue with event creation in event service layer/db ops", http.StatusFailedDependency)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&event)
}

// GET ::  /api/events/{eventUUID}
func (eh *EventHandler) GetEventDet(w http.ResponseWriter, r *http.Request) {
	eh.logger.Println("Get Event details for a single event")

	vars := mux.Vars(r)

	ctx := r.Context()

	eventIDStr := vars["eventUUID"]

	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		log.Fatalf("failed to convert string to uuid : 5v", err)
	}

	reqUSerID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		eh.logger.Fatalf("Failed to fetch userUUID from request context: %v", err)
		http.Error(w, "request context failed to provide USERID", http.StatusExpectationFailed)
		return
	}

	eventDetails, err := eh.es.GetTeamEvents(ctx, eventID, reqUSerID)
	if err != nil {
		log.Println("check service layer transactions-> there is an issue there")
		http.Error(w, "issue with get teams details function in service layer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&eventDetails)

}

func (eh *EventHandler) UpdateEventDetails(w http.ResponseWriter, r *http.Request) {
	eh.logger.Println("updating team event Details")

	ctx := r.Context()

	vars := mux.Vars(r)

	eventIDStr := vars["eventID"]
	teamIDStr := vars["userID"]
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		log.Fatalf("error changing to uuid")
	}

	reqUserID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "error fetching userID from context", http.StatusInternalServerError)
		return
	}

	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		log.Fatalf("issue changing user ID to uuid")
	}

	var update models.UpdateEventReq

	err = json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		http.Error(w, "issue decoding request data", http.StatusInternalServerError)
		return
	}

	updateEvent, err := eh.es.UpdateEventDetails(ctx, reqUserID, teamID, eventID, update)
	if err != nil {
		http.Error(w, "Error updating events from service layer", http.StatusExpectationFailed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&updateEvent)
}
