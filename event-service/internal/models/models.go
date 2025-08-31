package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID        uuid.UUID
	TeamID    uuid.UUID
	Title     string
	EventType string
	StartTime time.Time
	EndTime   time.Time
	Location  string
	Notes     string
	TeamName  string
}

type Attendance struct {
	EventID    uuid.UUID
	UserID     uuid.UUID
	Status     string
	UserName   string
	UpdateteAt time.Time
}
