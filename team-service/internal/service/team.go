package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github/wycliff-ochieng/internal/database"
	"github/wycliff-ochieng/internal/models"
	"log"
	"time"

	"github.com/wycliff-ochieng/sports-proto/user_grpc/user_proto"

	"github.com/google/uuid"
	//middleware "github.com/wycliff-ochieng/common_packages"
)

var ErrForbidden = errors.New("user not allowed here")
var ErrNotFound = errors.New("team not found/ does not exist")

type TeamService struct {
	db         database.DBInterface
	userClient user_proto.UserServiceRPCClient
}

type updateTeamReq struct {
	TeamID      uuid.UUID `json:"teamid"`
	Name        string    `json:"name"`
	Sport       string    `json:"sport"`
	Description string    `json:"description"`
	Createdat   time.Time `json:"createdat"`
	Updatedat   time.Time `json:"updatedat"`
}

func NewTeamService(db database.DBInterface, userClient user_proto.UserServiceRPCClient) *TeamService {
	return &TeamService{
		db:         db,
		userClient: userClient,
	}
}

// POST :: creating team /api/team/create
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

// GET::  all teams for a single user (pass userID)
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

// single team for a single user  ->  change this to repo service - > team details
func (ts *TeamService) GetTeamByID(ctx context.Context, teamID uuid.UUID) (*models.Team, error) {
	var AllTeams models.Team
	query := `SELECT * FROM teams WHERE team_id=$1`
	err := ts.db.QueryRowContext(ctx, query, teamID).Scan(&AllTeams)
	if err != nil {
		return nil, err
	}
	/*defer rows.Close()

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
	}*/
	return &AllTeams, err
}

// repo service to check if user is a member of a tema
func (ts *TeamService) IsTeamMember(ctx context.Context, userID uuid.UUID, teamID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM team_members WHERE user_id = $1 AND team_id = $2)`

	err := ts.db.QueryRowContext(ctx, query, userID, teamID).Scan(&exists)
	if err != nil {
		log.Fatalf("Error: query row transaction failed due to: %v", err)
	}
	return exists, nil
}

// repo service
func (ts *TeamService) GetTeamsMembers(ctx context.Context, teamID uuid.UUID) ([]*models.TeamMembers, error) {
	var teamMembers []*models.TeamMembers
	query := `SELECT * FROM team_members WHERE team_id=$1`
	rows, err := ts.db.QueryContext(ctx, query, teamID)
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

// GET :: get a single team by ID
func (ts *TeamService) GetTeamDetails(ctx context.Context, reqUserID uuid.UUID, teamID uuid.UUID) (*models.TeamDetailsInfo, error) {

	//is the user a member - > authorization check -> are you team member
	isTeamMember, err := ts.IsTeamMember(ctx, reqUserID, teamID)
	if err != nil {
		log.Fatalf("Error cecking membership: %v", err)
		return nil, err
	}

	if !isTeamMember {
		log.Fatal("Not Team Member :Not allowed to get team details")
		return nil, ErrForbidden
	}

	//get the all teams details for this teamID -> get teamsByID
	team, err := ts.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	allTeamMembers, err := ts.GetTeamsMembers(ctx, teamID)
	if err != nil {
		log.Fatal("No team members")
		return nil, err
	}

	if len(allTeamMembers) == 0 {
		return &models.TeamDetailsInfo{
			TeamID:      team.TeamID,
			Name:        team.Name,
			Sport:       team.Sport,
			Description: team.Description,
			Members:     []models.TeamMembers{},
		}, nil
	}

	//fetch all members uuid and their roles -> getTeamMmebers
	var memberUUID []string
	for _, member := range allTeamMembers {
		//mbrUUID,err := uuid.Parse(member.UserID)
		memberUUID = append(memberUUID, member.UserID.String())
	}

	//gRPC call to user service
	// TODO :: --> server grpc call client (user-service) to check if
	profilesReq := &user_proto.GetUserRequest{
		Userid: memberUUID,
	}

	profileRes, err := ts.userClient.GetUserProfiles(ctx, profilesReq)
	if err != nil {
		//handle error
		return nil, fmt.Errorf("could not fetch profiles from user service due to: %v", err)
	}

	userProfilesMap := profileRes.Profiles

	//gather final reponse struct
	finalResponse := models.TeamDetailsInfo{
		TeamID:      team.TeamID,
		Name:        team.Name,
		Sport:       team.Sport,
		Description: team.Description,
		Members:     make([]models.TeamMembers, 0, len(allTeamMembers)),
		//Joinedat: team.Createdat,
		//Updatedat: team.Updatedat,
	}

	//combine data from database to the final response

	for _, member := range allTeamMembers {
		profile, found := userProfilesMap[member.UserID.String()]
		if !found {
			log.Println("Warning , team member does no exists in the system")
			continue
		}

		UserUUID, err := uuid.Parse(profile.Userid)
		if err != nil {
			return nil, err
		}

		//finalResponse.Members = append(finalResponse.Members, models.TeamDetailsInfo{
		//	TeamID:   team.TeamID,
		//	UserID:   UserUUID,
		//	Role:     profile.Email,
		//	Joinedat: profile.Createdat,
		//})
		finalResponse.Members = append(finalResponse.Members, models.TeamMembers{
			TeamID:   member.TeamID,
			UserID:   UserUUID,
			Role:     member.Role,
			Joinedat: member.Joinedat,
		})
	}
	return &finalResponse, nil
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
	//teamID,err := ts.GetTeamByID()
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

	//TODO LATER :: produce the updateTeam Event to a Kafka topic  -> High Priority

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

// repo service for udating team
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

	err := ts.db.QueryRowContext(ctx, query, teamID, userID).Scan(&role)
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

	/*reqUserID, err := middleware.GetUserUUIDFromContext(ctx)
	if err != nil{
		return nil,err
	}*/

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
	//profiles,err :=

	//i've assumed the user exists in the system
	addedMember, err := ts.AddMember(ctx, teamID, addMember)
	if err != nil {
		log.Fatal("Error: Failed to add Member to team due to: ")
		return nil, err
	}

	if err := txs.Commit(); err != nil {
		log.Fatal("Error commit the transaction due to :")
		return nil, err
	}
	return addedMember, nil
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

func (ts *TeamService) GetTeamMebers(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (*models.TeamMembersResponse, error) {

	isMember, err := ts.IsTeamMember(ctx, userID, teamID)
	if err != nil {
		log.Fatalf("Error checking membership due to: %v", err)
	}

	if !isMember {
		return nil, ErrForbidden
	}

	query := `SELECT user_id, roles , joinedat FROM team_members WHERE teamid=$1`

	rows, err := ts.db.QueryContext(ctx, query, teamID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var TeamMembers []models.TeamMembers

	for rows.Next() {
		var members models.TeamMembers

		err := rows.Scan(
			&members.UserID,
			&members.Role,
			&members.Joinedat,
		)
		if err != nil {
			log.Fatalf("error scanning rows due to: %v", err)
		}

		TeamMembers = append(TeamMembers, members)

	}

	//grpc call to get user profiles

	return nil, nil
}
