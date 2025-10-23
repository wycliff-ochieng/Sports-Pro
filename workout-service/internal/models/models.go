package models

import (
	"time"

	"github.com/google/uuid"
)

type Workout struct {
	ID          uuid.UUID
	Name        string
	Description string
	Category    string
	CreatedBy   uuid.UUID //user_id of the coach / admin
	CreatedOn   time.Time
	UpdatedON   time.Time
}

type Exercise struct {
	ID          uuid.UUID
	Name        string
	Description string
	Instruction string
	CreatedBy   uuid.UUID
	CreatedOn   time.Time
	UpdatedOn   time.Time
}

type WorkoutExerciseResponse struct {
	WorkoutID   uuid.UUID
	Name        string
	ExcerciseID uuid.UUID
	Order       int32
	Sets        int32
	Reps        int32
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
