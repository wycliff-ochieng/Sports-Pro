package service

import (
	"context"
	"github/wycliff-ochieng/internal/database"
	"github/wycliff-ochieng/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
)

type TeamService struct {
	db database.DBInterface
}

func NewTeamService(db database.DBInterface) *TeamService {
	return &TeamService{db}
}

func (ts *TeamService) CreateTeam(ctx context.Context, teamID uuid.UUID, name string, sport string, description string, createdat, updatedat time.Time) (*models.Team, error) {

	var team *models.Team

	team, err := models.NewTeam(teamID, name, sport, description, createdat, updatedat)
	if err != nil {
		log.Fatalf("error creating new team")
	}

	query := `INSERT INTO teams(id,name,sports,description,createdat,updatedat) VALUES($1,$2,$3,$4,$5,$6)`

	_, err = ts.db.ExecContext(ctx, query, team.TeamID, team.Name, team.Sport, team.Description, team.Createdat, team.Updatedat)
	if err != nil {
		log.Printf("ERROR creating team due to: %v", err)
	}

	return &models.Team{
		TeamID:      teamID,
		Name:        name,
		Sport:       sport,
		Description: description,
		Createdat:   createdat,
	}, nil
}

func (ts *TeamService) GetMyTeams(ctx context.Context, userID uuid.UUID) (*[]models.Team, error) {

	//get userUUID from the context
	return nil, nil
}
