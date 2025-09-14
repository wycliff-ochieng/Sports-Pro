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
	EventID uuid.UUID
	TeamID  uuid.UUID
	UserID  uuid.UUID
	Status  string
	//UserName   string
	UpdateteAt time.Time
}

type CreateEventReq struct {
	EventID   uuid.UUID
	TeamID    uuid.UUID
	Name      string
	EventType string
	Location  string
	StartTime time.Time
	EndTime   time.Time
}

type AttendanceResponse struct {
	UserID      uuid.UUID
	FirstName   string
	LastName    string
	EventStatus string
	Email       string
}

type EventDetails struct {
}

func NewEvent(teamID uuid.UUID, name string, eventype string, Location string, start, end time.Time) (*Event, error) {
	return &Event{
		TeamID:    teamID,
		Title:     name,
		EventType: eventype,
		Location:  Location,
		StartTime: start,
		EndTime:   end,
	}, nil
}

func NewAttendance(eventID uuid.UUID, teamID uuid.UUID, userID uuid.UUID, status string, updatedat time.Time) (*Attendance, error) {
	return &Attendance{
		EventID:    eventID,
		TeamID:     teamID,
		UserID:     userID,
		Status:     status,
		UpdateteAt: updatedat,
	}, nil
}
