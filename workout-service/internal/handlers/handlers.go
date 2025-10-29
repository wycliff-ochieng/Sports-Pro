package handlers

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strconv"

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
	Category    string
	Exercises   []models.WorkoutExerciseResponse
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
	workout, err := h.ws.CreateWorkout(ctx, userID, WorkOutReq.Name, WorkOutReq.Category, WorkOutReq.Description, WorkOutReq.Exercises)
	if err != nil {
		http.Error(w, "issue creating workout in service layer", http.StatusExpectationFailed)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&workout)
}

func (h *WorkoutHandler) GetAllWorkouts(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get all workout present")

	ctx := r.Context()

	parameters := r.URL.Query()
	limitString := parameters.Get("limit")
	cursor := parameters.Get("cursor")
	search := parameters.Get("search")

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		log.Printf("error converting limit to integer, %s", err)
	}

	minLimit := 1
	maxLimit := 100
	defaultLimit := 25

	//get userID from context

	if limit < minLimit || limit >= maxLimit {
		http.Error(w, "Limit is either less or greater than the limit", http.StatusFailedDependency)
		return
	}

	if limit == 0 {
		limit = defaultLimit
	}

	createdParams := models.ListWorkoutParams{
		Search: search,
		Limit:  limit,
		Cursor: cursor,
	}

	userID, err := 

	//call service layer for all the workouts
	workouts, err := h.ws.ListAllWorkouts()

	//
}
