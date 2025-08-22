package service

import (
	"context"
	"database/sql"
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

type updateTeamReq struct {
	TeamID      uuid.UUID `json:"teamid"`
	Name        string    `json:"name"`
	Sport       string    `json:"sport"`
	Description string    `json:"description"`
	Createdat   time.Time `json:"createdat"`
	Updatedat   time.Time `json:"updatedat"`
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

// teams for a single user  ->  change this to repo service
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

// get team ID/ UUID
func (ts *TeamService) GetTeamByUUID(tcx context.Context) {}

func (ts *TeamService) GetTeamDetails(ctx context.Context, userID uuid.UUID) {

	//is the user a member - > authorization check -> are you team member

	//get the all teams details for this teamID -> get teamsByID

	//fetch all members uuid and their roles -> getTeamMmebers
}

func (ts *TeamService) UpdateTeamDetails(ctx context.Context, teamID uuid.UUID, reqUserID uuid.UUID, updateData models.UpdateTeamReq) (*models.Team, error) {
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
	role, err := ts.GetRoleForUser(ctx, teamID, reqUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("user is not a memeber of this team")
		}
		return nil, err
	}
	//check roles
	isUserAuthorized := role == "COACH" || role == "MANAGER"
	if !isUserAuthorized {
		log.Fatal("ERROR : user is not Allowed to edit")
	}

	//if they  are authorized: Database write operation
	updatedTeam, err := ts.UpdateTeam(ctx, teamID, updateData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		log.Fatal("Fialed to update team deatils")
		return nil, err
	}

	//commit the transactions
	if err := txs.Commit(); err != nil {
		log.Fatalf("Failed to commit the transaction, rollback wikk be initiated: %v", err)
	}

	//TODO LATER :: produce the updateTeam Event to a Kafka topic

	return updatedTeam, nil
}

/*func (ts *TeamService) isTeam(ctx context.Context, teamID uuid.UUID) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM team WHERE team_id = $1)`
	err := ts.db.QueryRowContext(ctx, query).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, err
}
*/

func (ts *TeamService) UpdateTeam(ctx context.Context, teamID uuid.UUID, updateData models.UpdateTeamReq) (*models.Team, error) {
	var team models.Team
	query := `UPDATE teams SET name=$1, description=$, updateat=Now() WHERE team_id=$3
	RETURNING name,sport,description,createdat,updatedat`

	err := ts.db.QueryRowContext(ctx, query, updateData.Name, updateData.Description, teamID).Scan(&team.TeamID, &team.Name, &team.Sport, &team.Description, &team.Createdat, &team.Updatedat)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func (ts *TeamService) GetRoleForUser(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (string, error) {
	var role string

	query := `SELECT FROM team_members WHERE team_id = $1 AND user_id=$2`

	err := ts.db.QueryRowContext(ctx, query).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

// Add a memeber to a team -> POST
func (ts *TeamService) AddTeamMember(ctx context.Context, teamID uuid.UUID, reqUserID uuid.UUID, addMember models.AddMemberReq) (*models.TeamMembers, error) {

	txs, err := ts.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer txs.Rollback()

	//query team memebr to get role -> Auhtorization check
	role, err := ts.GetRoleForUser(ctx, teamID, reqUserID)
	if err != nil {
		return nil, err
	}

	isAllowed := role == "COACH" || role == "MANAGER"
	if !isAllowed {
		return nil, ErrForbidden
	}

	//check if user being added to the team exists in the system
	//Will need to make a gRPC call to user-service
	//TODO:: - > implementing gRPC communication

	//i've assumed the user exists in the system
	addedMember, err := ts.AddMember()
	return nil, nil
}

func (ts *TeamService) AddMember(ctx context.Context, teamID uuid.UUID, addedMember models.AddMemberReq) (*models.TeamMembers, error) {
	var member models.TeamMembers
	query := `INSERT INTO team_members(team_id,role,joinedat,user_id,) VALUES($1,$2,$3,$4)`

	_, err := ts.db.ExecContext(ctx, query, member.TeamID, member.Role, member.Joinedat, member.UserID)
	if err != nil {
		return nil, err
	}

	return &models.TeamMembers{
		TeamID:   teamID,
		Role:     addedMember.Role,
		Joinedat: addedMember.Joinedat,
		UserID:   addedMember.UserID,
	}, nil
}
