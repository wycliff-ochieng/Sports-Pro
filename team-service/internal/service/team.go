package service

import (
	"context"
	"github/wycliff-ochieng/internal/database"
	"github/wycliff-ochieng/internal/models"
	"log"
	"time"

	//"github.com/wycliff-ochieng/common_packages"

	"github.com/google/uuid"
	//middleware "github.com/wycliff-ochieng/common_packages"
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

func (ts *TeamService) GetMyTeams(ctx context.Context, userID uuid.UUID) (*[]models.TeamInfo, error) {

	var teams []models.TeamInfo
	//	var members models.TeamMembers

	query := `SELECT t.id,t.name,t.sport,t.description,t.createdat,tm.Role,tm.joined FROM teams t LEFT JOIN team_members ON tm.user_id = $1 WHERE t.id = tm.team_id`

	rows, err := ts.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Fatalf("failed to fetch teams for %s bacause of %s", err, userID)
	}
	defer rows.Close()

	for rows.Next() {
		//var myTeams models.Team
		var myTeams models.TeamInfo

		err := rows.Scan(
			&myTeams.TeamID,
			&myTeams.Name,
			&myTeams.Sport,
			&myTeams.Role,
			&myTeams.Description,
			&myTeams.Updatedat,
			&myTeams.Joinedat,
		)
		if err != nil {
			log.Fatalf("Failed to loop through all teams : %v", err)
		}
		teams = append(teams, myTeams)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &teams, nil
}
