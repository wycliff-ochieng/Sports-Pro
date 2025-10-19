package handlers

import (
	"log/slog"
	"net/http"

	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/internal/service"
	auth "github.com/wycliff-ochieng/sports-common-package"
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

	ctx := r.Context()

	userID, err := auth.Get

	return
}
