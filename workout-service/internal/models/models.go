package models

import (
	"time"

	"github.com/google/uuid"
)

type Workout struct {
	ID          uuid.UUID
	Name        string
	Category    string
	Description string
	CreatedBy   uuid.UUID //user_id of the coach / admin
	CreatedOn   time.Time
	UpdatedON   time.Time
}

type Exercise struct {
	ID    uuid.UUID
	Name  string
	Order int
	Sets  int
	Reps  int
	//Description string
	//Instruction string
	//CreatedBy   uuid.UUID
	CreatedOn time.Time
	UpdatedOn time.Time
}

type WorkoutExerciseResponse struct {
	WorkoutID uuid.UUID
	//Name        string
	//Description string
	ExcerciseID uuid.UUID
	Order       int32
	Sets        int32
	Reps        int32
}

type WorkoutExercise struct{}

type CreateWorkoutResponse struct {
	WorkoutID   uuid.UUID
	Name        string
	Category    string
	Description string
	Createdat   time.Time
	Exercises   []Exercise
}

type Media struct {
	ID             uuid.UUID
	ParentID       uuid.UUID
	PArentType     string
	StorageProvide string
	BucketName     string
	ObjectKey      string
	Filename       string
	MimeType       string
	UploadedAT     time.Time
}

type ListWorkoutParams struct {
	Search string
	Limit  int
	Cursor string
}

type PaginatedWorkout struct {
	Data       []Workout
	NextCursor string
}
