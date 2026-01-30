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
	TeamID    uuid.UUID `json:"teamid"`
	Role      string    `json:"role"`
	Joinedat  time.Time `json:"joinedat"`
	UserID    uuid.UUID `json:"userid"`
	Firstname string    `json:"firstName,omitempty"`
	Lastname  string    `json:"lastName,omitempty"`
	Email     string    `json:"email,omitempty"`
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

type TeamMembersResponse struct {
	UserID    uuid.UUID
	Firstname string `json:"firstName"`
	Lastname  string `json:"lastName"`
	Email     string
	Createdat time.Time
	Updatedat time.Time
}

type TeamDetailsInfo struct {
	TeamID      uuid.UUID
	Name        string
	Sport       string
	Description string
	Members     []TeamMembers
}

type UpdateTeamReq struct {
	TeamID      uuid.UUID `json:"teamid"`
	Name        string    `json:"name"`
	Sport       string    `json:"sport"`
	Description string    `json:"description"`
	Createdat   time.Time `json:"createdat"`
	Updatedat   time.Time `json:"updatedat"`
}

type AddMemberReq struct {
	UserID     uuid.UUID `json:"-"` // resolved target user UUID
	Role       string    `json:"role"`
	Joinedat   time.Time `json:"joinedat"`
	Identifier string    `json:"userid"` // accepts uuid or email from client
}

type UpdateTeamMemberReq struct {
	Role string
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

//func NewTeamInfo() ([]*TeamInfo,error){
//	return []&TeamInfo{},nil
//}
