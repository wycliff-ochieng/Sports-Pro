package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	TeamID      uuid.UUID
	Name        string
	Sport       string
	Description string
	Createdat   time.Time
	Updatedat   time.Time
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
