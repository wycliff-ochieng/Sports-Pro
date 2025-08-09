package service

import (
	"github/wycliff-ochieng/internal/database"
	"github/wycliff-ochieng/internal/models"
	"time"

	"github.com/google/uuid"
)

type TeamService struct {
	db database.DBInterface
}

func NewTeamService(db database.DBInterface) *TeamService {
	return &TeamService{db}
}

func (ts *TeamService) CreateTeam(id uuid.UUID, name string, sport string, description string, createdat time.Time, updatedat time.Time) (*models.Team, error) {
	return nil, nil
}
