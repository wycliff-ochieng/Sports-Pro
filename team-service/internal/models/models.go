package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	TeamID      uuid.UUID `json:"temaid"`
	Name        string    `json:"name"`
	Sport       string    `json:"sport"`
	Description string    `json:"description"`
	Createdat   time.Time `json::"createdat"`
	Updatedat   time.Time `json:"updatedat"`
}

type TeamMembers struct {
	TeamID   uuid.UUID
	UserID   uuid.UUID
	Role     string
	Joinedat time.Time
}

func NewTeam(teamID uuid.UUID, name string, sport string, description string, createdat time.Time, updatedat time.Time) *Team {
	return &Team{
		TeamID:      teamID,
		Name:        name,
		Sport:       sport,
		Description: description,
		Createdat:   createdat,
		Updatedat:   updatedat,
	}
}

func NewTeamMembers(teamID uuid.UUID, userID uuid.UUID, role string, joinedat time.Time) *TeamMembers {
	return &TeamMembers{
		TeamID:   teamID,
		UserID:   userID,
		Role:     role,
		Joinedat: joinedat,
	}
}
