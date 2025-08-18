package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	TeamID      uuid.UUID `json:"teamid"`
	Name        string    `json:"name"`
	Sport       string    `json:"sport"`
	Description string    `json:"description"`
	Createdat   time.Time `json:"createdat"`
	Updatedat   time.Time `json:"updatedat"`
}

type TeamMembers struct {
	TeamID   uuid.UUID `json:"teamid"`
	UserID   uuid.UUID `json:"userid"`
	Role     string    `json:"role"`
	Joinedat time.Time `json:"joinedat"`
}

type TeamInfo struct {
	TeamID      uuid.UUID
	Name        string
	Sport       string
	Role        string
	Description string
	Updatedat   time.Time
	Joinedat    time.Time
}

func NewTeam(teamID uuid.UUID, name string, sport string, description string, createdat, updatedat time.Time) (*Team, error) {
	return &Team{
		TeamID:      uuid.New(),
		Name:        name,
		Sport:       sport,
		Description: description,
		Createdat:   time.Now().UTC(),
		Updatedat:   time.Now().UTC(),
	}, nil
}

func NewTeamMembers(teamID uuid.UUID, userID uuid.UUID, role string, joinedat time.Time) *TeamMembers {
	return &TeamMembers{
		TeamID:   teamID,
		UserID:   userID,
		Role:     role,
		Joinedat: joinedat,
	}
}
