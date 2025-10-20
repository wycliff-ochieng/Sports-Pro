package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/internal/service"
	auth "github.com/wycliff-ochieng/sports-common-package/middleware"
)

type WorkoutHandler struct {
	logger *slog.Logger
	ws     *service.WorkoutService
}

type CreateWorkoutReq struct {
	Name        string
	Description string
	Exercises   []models.Exercise
}

func NewWorkoutHandler(logger *slog.Logger, ws *service.WorkoutService) *WorkoutHandler {
	return &WorkoutHandler{
		logger: logger,
		ws:     ws,
	}
}

func (h *WorkoutHandler) CreateWorkout(w http.ResponseWriter, r *http.Request) {
	//must be coach/ admin -> RequireRole middleware
	//posting video, image
	//must be authenticated -> check user_id from context

	//http.ServeFile()

	var WorkOutReq CreateWorkoutReq

	ctx := r.Context()

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "issue with getting userID from middleware", http.StatusInternalServerError)
		return
	}

	//request body
	if err := json.NewDecoder(r.Body).Decode(&WorkOutReq); err != nil {
		http.Error(w, "issue with unmarshaling the incoming request", http.StatusInternalServerError)
		return
	}

	//validation
	if WorkOutReq.Name == " " || WorkOutReq.Description == " " || len(WorkOutReq.Exercises) == 0 {
		http.Error(w, "please enter valid data", http.StatusExpectationFailed)
		return
	}

	//call service layer

	return
}
