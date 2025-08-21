package service

import (
	"context"
	"errors"
	"github/wycliff-ochieng/internal/database"
	"github/wycliff-ochieng/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
	//middleware "github.com/wycliff-ochieng/common_packages"
)

var ErrForbidden = errors.New("user not allowed here")
var ErrNotFound = errors.New("team not found/ does not exist")

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

// teams for a single user
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

// teams for a single user  ->  hange this to repo service
func (ts *TeamService) GetTeamByID(ctx context.Context, teamID string, userID uuid.UUID) ([]*models.Team, error) {
	var AllTeams []*models.Team
	query := `SELECT * FROM teams WHERE user_id=$1`
	rows, err := ts.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var allTeams models.Team

		err := rows.Scan(
			&allTeams.Name,
			&allTeams.Sport,
			&allTeams.Description,
			&allTeams.TeamID,
			&allTeams.Createdat,
			&allTeams.Updatedat,
		)
		if err != nil {
			log.Fatalf("Error: scnning rows issue due to: %v", err)
		}
		AllTeams = append(AllTeams, &allTeams)
	}
	return AllTeams, err
}

// repo service
func (ts *TeamService) IsTeamMember(ctx context.Context, userID uuid.UUID, teamID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM team_members WHERE user_id = $1 AND team_id = $2)`

	err := ts.db.QueryRowContext(ctx, query).Scan(&exists)
	if err != nil {
		log.Fatalf("Error: query row transaction failed due to: %v", err)
	}
	return exists, nil
}

// repo service
func (ts *TeamService) GetTeamsMembers(ctx context.Context, teamID uuid.UUID) ([]*models.TeamMembers, error) {
	var teamMembers []*models.TeamMembers
	query := `SELECT * FROM team_members WHERE team_id=$1`
	rows, err := ts.db.QueryContext(ctx, query)
	if err != nil {
		log.Fatalf("Error: issue with fetchiing team members: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var members models.TeamMembers

		err := rows.Scan(
			&members.TeamID,
			&members.UserID,
			&members.Role,
			&members.Joinedat,
		)
		if err != nil {
			log.Fatalf("Error scanning rows: %v", err)
		}
		teamMembers = append(teamMembers, &members)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Team has no memebers :%v", err)
	}
	return teamMembers, nil
}

func (ts *TeamService) GetTeamDetails(ctx context.Context, userID uuid.UUID) {

	//is the user a member - > authorization check

	//get the all teams details for this teamID

	//fetch all members uuid and their roles
}

func (ts *TeamService) UpdateTeamDetails(ctx context.Context, name string, description string) (*models.TeamInfo, error) {
	//check roles ->is the user a coach /manager of that team -> RBAC
	//roles, err := middleware.GetUserRoleFromContext(ctx)
	//if err != nil {
	//	return nil, err //log.Fatalf("FATAL. NOT ALLOWED: No role found for this user: %v", err)
	//}

	txs, err := ts.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer txs.Rollback()

	//check if team exists
	team, err := ts.db.GetTeamByID()

	//

	return nil, nil
}
