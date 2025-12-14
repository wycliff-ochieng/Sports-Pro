package handlers

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	//"github.com/gofrs/uuid"
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
	Exercises   []models.ExerciseInCreateWorkoutReq
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

	var WorkOutReq models.CreateWorkoutResponse

	ctx := r.Context()

	//request body
	if err := json.NewDecoder(r.Body).Decode(&WorkOutReq); err != nil {
		http.Error(w, "issue with unmarshaling the incoming request", http.StatusInternalServerError)
		return
	}

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		log.Printf("due to : %s", err)
		http.Error(w, "issue with getting userID from middleware", http.StatusInternalServerError)
		return
	}

	//validation
	if WorkOutReq.Name == " " || WorkOutReq.Description == " " || len(WorkOutReq.Exercises) == 0 {
		http.Error(w, "please enter valid data", http.StatusExpectationFailed)
		return
	}

	//call service layer
	workout, err := h.ws.CreateWorkout(ctx, userID, WorkOutReq) //WorkOutReq.Category, WorkOutReq.Description, WorkOutReq.Exercises)
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
		http.Error(w, "conversion error:", http.StatusExpectationFailed)
		return
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

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "error getting userID from ctx", http.StatusFailedDependency)
		return
	}

	//call service layer for all the workouts
	workouts, err := h.ws.ListAllWorkouts(ctx, userID, createdParams)
	if err != nil {
		http.Error(w, "error getting workouts in service layer", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&workouts)

	//
}

func (h *WorkoutHandler) CreateExercise(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Create Exercise Handler now in actions")

	ctx := r.Context()

	var exercise models.CreateExerciseReq

	if err := json.NewDecoder(r.Body).Decode(&exercise); err != nil {
		http.Error(w, "Error decoding request body", http.StatusInternalServerError)
		return
	}

	//validations-> make sure name is not empty

	if exercise.Name == "" {
		http.Error(w, "enter valid exercise name", http.StatusExpectationFailed)
		return
	}

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "issue with getting userID from context", http.StatusInternalServerError)
		return
	}

	//call service layer

	ex, err := h.ws.CreateExercise(ctx, userID, exercise.Name, exercise.Description, exercise.Instructions)
	if err != nil {
		http.Error(w, "some error in service layer during creation", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&ex)

}

func (h *WorkoutHandler) GetAllExercises(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	//get user ID

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "issue getting exercises in service layer", http.StatusInternalServerError)
		return
	}

	exercises, err := h.ws.GetExercises(ctx, userID)
	if err != nil {
		http.Error(w, "error getting exercises from db", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&exercises)
}

func (h *WorkoutHandler) GetWorkotDetail(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Get a single Workout detail page")
	return
}

func (h *WorkoutHandler) GetExerciseDetail(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Geting an exercise by specific ID")
	return
}

func (h *WorkoutHandler) MediaPresignedURL(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Media upload temporary URL")

	ctx := r.Context()

	//send file metadata not file

	var Metadata *models.PresignedURLReq

	err := json.NewDecoder(r.Body).Decode(&Metadata)
	if err != nil {
		http.Error(w, "Issue decoding URL metadata", http.StatusInternalServerError)
		return
	}

	userID, err := auth.GetUserUUIDFromContext(ctx)
	if err != nil {
		http.Error(w, "failed to get uuid from contextx", http.StatusFailedDependency)
		return
	}

	var AllowedContent = []string{"image/jpeg", "image/png"}

	if Metadata.ParentType != "workout" && Metadata.ParentType != "exercise" {
		http.Error(w, "invalid parent type", http.StatusBadRequest)
		return
		//if Metadata.MimeType == AllowedContent[]
	}

	isAllowed := false
	for _, allowed := range AllowedContent {
		if Metadata.MimeType == allowed {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		http.Error(w, "Please upload the required content type", http.StatusExpectationFailed)
		return
	}

	//call service layer
	presignedURL, err := h.ws.GeneratePresignedURL(ctx, userID, Metadata)
	if err != nil {
		http.Error(w, "issue generating presigned url in service layer", http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&presignedURL)
}

func (h *WorkoutHandler) UploadComplete(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Media upload complete handler")

	ctx := r.Context()

	var req models.MediaUploadCompleteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Issue decoding request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if req.ParentID == uuid.Nil || req.ObjectKey == "" || req.Filename == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Call service
	media, err := h.ws.SaveMediaMetadata(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to save media metadata", "error", err)
		http.Error(w, "Failed to save media metadata", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(media)
}
