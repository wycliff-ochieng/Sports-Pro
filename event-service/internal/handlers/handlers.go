package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
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
