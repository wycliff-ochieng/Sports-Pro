package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/sports-common-package/team_grpc/team_proto"
	"github.com/wycliff-ochieng/sports-common-package/user_grpc/user_proto"
)

var (
	ErrForbidden = errors.New("Not allowed")
	ErrNotFound  = errors.New("Not found in system")
)

type Events interface {
	//CreateTeamEvent(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID, eventTitle string, eventType string, location string, startTime time.Time, endTime time.Time) (*models.Event, error)
	GetTeamEvents()
	GetEventByID()
	UpdateEventDetails()
	CancelTeamEvent()
}

type EventService struct {
	db         database.DBInterface
	teamClient team_proto.TeamRPCClient
	userClient user_proto.UserServiceRPCClient
	l          *slog.Logger
}

func NewEventService(db database.DBInterface, teamClient team_proto.TeamRPCClient, userCllient user_proto.UserServiceRPCClient, logger *slog.Logger) *EventService {
	return &EventService{
		db:         db,
		teamClient: teamClient,
		userClient: userCllient,
		l:          logger,
	}
}

func (es *EventService) CreateTeamEvent(ctx context.Context, reqUserID uuid.UUID, eventID uuid.UUID, eventTitle string, teamID uuid.UUID, eventType string, location string, startTime time.Time, endTime time.Time) (*models.Event, error) {
	//Authorization via gRPC
	es.l.Info("Creation of event initiated by user")

	//var CreateEventReq models.CreateEventReq

	membershipReq := &team_proto.GetTeamMembershipRequest{
		TeamId: teamID.String(),
		UserId: []string{reqUserID.String()},
	}

	//log.Printf("T-ID: %s", CreateEventReq.TeamID)

	members, err := es.teamClient.CheckTeamMembership(ctx, membershipReq)
	if err != nil {
		return nil, err
	}

	log.Printf("User: %s", reqUserID)
	log.Printf("TeamID: %s", teamID)

	//check response/error from grRPC call- > if user is a member of that team or not and if member is coach/manager
	membersInfo, found := members.Members[reqUserID.String()]
	if !found {
		es.l.Error("user does not exist in the team therefore is not a member")
		return nil, ErrForbidden
	}

	isAuthorized := membersInfo.Role == "coach" || membersInfo.Role == "manager"
	if !isAuthorized {
		es.l.Warn("user is not a coach/manager therefore cant create events")
		//return nil, err
	}
	log.Printf("Role: %s", membersInfo.Role)
	es.l.Info("Authorization is successfull")

	//get team data for attendance table insert(business logic)
	teamMembersReq := team_proto.GetTeamSummaryRequest{TeamId: teamID.String()}
	teamMembersRes, err := es.teamClient.GetTeamSummary(ctx, &teamMembersReq)
	if err != nil {
		es.l.Error("gRPC call to team service failed")
		return nil, fmt.Errorf("cause of failure: %v", err)
	}

	teamMembers := teamMembersRes.Members

	//database atomic transaction ->

	//start transactions
	txs, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		es.l.Error("Issue with transaction start up")
		return nil, err
	}

	defer txs.Rollback()

	/*newEvent := &models.Event{
		TeamID:    CreateEventReq.TeamID,
		Title:     CreateEventReq.Name,
		EventType: CreateEventReq.EventType,
		Location:  CreateEventReq.Location,
		StartTime: CreateEventReq.StartTime,
		EndTime:   CreateEventReq.EndTime,
	}*/

	createdEvent, err := es.CreateEvent(ctx, txs, teamID, eventTitle, eventType, location, startTime, endTime)
	if err != nil {
		es.l.Error("error creating event due to ", "error", err)
		return nil, err
	}
	es.l.Info("event created successfully for team", "teamID", teamID)
	es.l.Info("event", "createdEvent", createdEvent)

	//write to database the attendace list
	//prepopulate the initial list
	if len(teamMembers) > 0 {
		var attendanceRecords []models.Attendance
		for _, member := range teamMembers {
			memberID, err := uuid.Parse(member.UserId)
			if err != nil {
				return nil, err
			}

			attendanceRecords = append(attendanceRecords, models.Attendance{
				EventID:    createdEvent.ID,
				UserID:     memberID,
				TeamID:     teamID,
				Status:     "PENDING",
				UpdateteAt: time.Now(),
			})
		}

		log.Printf("members: %s", attendanceRecords)

		//insert into attendance table (bulk insert->provides high performance)
		if err := es.CreateBulkAttendance(ctx, txs, attendanceRecords); err != nil {
			es.l.Error("error inserting to iniital attendance table")
			return nil, err
		}
		es.l.Info("bulk rcord attendance created ")

		if err := txs.Commit(); err != nil {
			log.Fatalf("error commiting /creating team event due to: %v", err)
			return nil, err
		}
		es.l.Info("successfully created event for team %s", "teamID", teamID)

	}

	return &models.Event{}, nil
}

func (es *EventService) CreateEvent(ctx context.Context, tx *sql.Tx, teamID uuid.UUID, name string, eventtype string, location string, starttime, endtime time.Time) (*models.Event, error) {
	es.l.Info("Create event database execution")

	var newEvent models.Event

	query := `INSERT INTO events(event_title,event_type,location,start_time,end_time) VALUES($1,$2,$3,$4,$5)
	RETURNING event_id,event_title, event_type, location, start_time, end_time`

	err := es.db.QueryRowContext(ctx, query, name, eventtype, location, starttime, endtime).Scan(
		&newEvent.ID,
		//&newEvent.TeamID,
		&newEvent.Title,
		&newEvent.EventType,
		&newEvent.Location,
		&newEvent.StartTime,
		&newEvent.EndTime,
	)
	if err != nil {
		log.Fatalf("Error executing query:%v", err)
		return nil, fmt.Errorf("issue inserting events ")
	}
	return &models.Event{
		ID:        newEvent.ID,
		Title:     newEvent.Title,
		EventType: newEvent.EventType,
		Location:  newEvent.Location,
		StartTime: newEvent.StartTime,
		EndTime:   newEvent.EndTime,
	}, err
}

// Renamed for clarity and purpose. 'Insert' is redundant.
func (es *EventService) CreateBulkAttendance(ctx context.Context, tx *sql.Tx, records []models.Attendance) error {
	if len(records) == 0 {
		return nil
	}

	// Use a strings.Builder for efficiency.
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`INSERT INTO attendance (event_id, user_id, team_id, event_status, updated_at) VALUES `)

	const columnCount = 5
	vals := make([]interface{}, 0, len(records)*columnCount)

	for i, record := range records {
		n := i * columnCount
		// FIX: Generate correct '$' placeholders.
		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5))

		if i < len(records)-1 {
			queryBuilder.WriteString(",")
		}

		// Make sure the values match the columns in the INSERT statement.
		vals = append(vals, record.EventID, record.UserID, record.TeamID, record.Status, record.UpdateteAt)
	}

	finalQuery := queryBuilder.String()

	// No need to Prepare; ExecContext is sufficient.
	_, err := tx.ExecContext(ctx, finalQuery, vals...)
	if err != nil {
		// Log and return a wrapped error.
		log.Printf("failed to execute bulk attendance insert: %v", err)
		return fmt.Errorf("failed to execute bulk attendance insert: %w", err)
	}

	return nil
}

func (es *EventService) GetTeamEvents(ctx context.Context, eventID uuid.UUID, reqUserID uuid.UUID) (*models.EventDetails, error) {
	es.l.Info("getting a single event details, ")

	query := `SELECT teamID FROM events WHERE eventID=$1`

	var teamID string

	err := es.db.QueryRowContext(ctx, query, eventID).Scan(&teamID)
	if err == sql.ErrNoRows {
		es.l.Error("No team associated with thatt event")
		return nil, err
	}

	//call get events
	event, err := es.GetEvent(ctx, eventID)
	if err != nil {
		es.l.Error("Error getting events from database to repository layer")
		log.Printf("The error: %v", err)
		return nil, err
	}

	//gRPC call to team service to check membership
	memberShipReq := &team_proto.GetTeamMembershipRequest{
		TeamId: event.TeamID.String(),
		UserId: []string{reqUserID.String()},
	}

	members, err := es.teamClient.CheckTeamMembership(ctx, memberShipReq)
	if err != nil {
		return nil, err
	}

	log.Printf("TeamID: %s", event.TeamID.String())

	membersInfo, found := members.Members[reqUserID.String()]
	if !found {
		es.l.Info("Check if the requesting userId is a member of this team == reqUserId and teamID match")
		es.l.Error("userID not is not a member")
		return nil, ErrForbidden
	}

	isAllowed := membersInfo.Role == "coach" || membersInfo.Role == "manager" || membersInfo.Role == "player"
	if !isAllowed {
		es.l.Error("user does not posses a role in this team")
	}
	es.l.Info("authorization done successfully")

	//call get event attendace
	attendanceList, err := es.GetAttendanceList(ctx, eventID)
	if err != nil {
		es.l.Error("")
		log.Fatalf("Error due to : %v", err)
		return nil, err
	}

	userIDs := make([]string, 0, len(attendanceList))
	for _, record := range attendanceList {
		userIDs = append(userIDs, record.UserID.String())
	}

	log.Printf("Attendees IDs: %s", userIDs)

	profileReq := &user_proto.GetUserRequest{
		Userid: userIDs,
	}

	profileRes, err := es.userClient.GetUserProfiles(ctx, profileReq)
	if err != nil {
		es.l.Info("batch fetching user profiles for attendance list and event detail enrichment")
		return nil, err
	}

	userProfilesMap := profileRes.Profiles

	var finalAttendanceList []models.AttendanceResponse

	finalAttendanceList = make([]models.AttendanceResponse, len(attendanceList))

	for _, attendee := range attendanceList {
		profile, found := userProfilesMap[attendee.UserID.String()]
		if !found {
			es.l.Error("No such user in user service")
			continue
		}

		UserIDUUID, err := uuid.Parse(profile.Userid)
		if err != nil {
			es.l.Error("cannot convert to uuid")
		}

		attendanceResp := models.AttendanceResponse{
			UserID:    UserIDUUID,
			FirstName: profile.Firstname,
			LastName:  profile.Lastname,
			Email:     profile.Email,
			//TeamID: attendee.EventID,
			EventStatus: attendee.Status,
		}

		finalAttendanceList = append(finalAttendanceList, attendanceResp)
	}

	enrichedList := models.EventDetails{
		EventID:    event.ID,
		TeamID:     event.TeamID,
		EventName:  event.Title,
		Location:   event.Location,
		StartTime:  event.StartTime,
		EndTime:    event.EndTime,
		Attendance: finalAttendanceList,
	}

	//var userIDs []string
	//for _, ID := range{}
	return &enrichedList, nil
}

func (es *EventService) GetAttendanceList(ctx context.Context, eventID uuid.UUID) ([]*models.Attendance, error) {
	es.l.Info("Fetching attendance list repository operation")

	var EventAttendance []*models.Attendance

	query := `SELECT user_id,event_status,updated_at FROM attendance WHERE event_id=$1`

	rows, err := es.db.QueryContext(ctx, query, eventID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var attendance models.Attendance
		err = rows.Scan(
			//&attendance.EventID,
			//&attendance.TeamID,
			&attendance.UserID,
			&attendance.Status,
			&attendance.UpdateteAt,
		)

		EventAttendance = append(EventAttendance, &attendance)
	}

	return EventAttendance, err
}

func (es *EventService) GetEvent(ctx context.Context, eventID uuid.UUID) (*models.Event, error) {
	es.l.Info(" Fetching the event detailed data database operations")

	var event models.Event

	//query := `SELECT event_id,event_title,event_type,location,start_time,end_time FROM events WHERE event_id = $1`
	query := `SELECT e.event_id,a.team_id,e.event_title,e.event_type,e.location,e.start_time,e.end_time FROM events e JOIN attendance a ON e.event_id = a.event_id`

	err := es.db.QueryRowContext(ctx, query).Scan(
		&event.ID,
		&event.TeamID,
		&event.Title,
		&event.EventType,
		&event.Location,
		&event.StartTime,
		&event.EndTime,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (es *EventService) UpdateEventDetails(ctx context.Context, reqUserID uuid.UUID, teamID uuid.UUID, eventID uuid.UUID, toUpdate models.UpdateEventReq) (*models.Event, error) {
	es.l.Info("PUT operation for the event service")

	txs, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		es.l.Error("Error while trying to begin transactions")
	}

	defer txs.Rollback()

	team, err := es.GetEvent(ctx, eventID)
	if err != nil {
		es.l.Error("failed to get event details")
		return nil, err
	}

	teamMembersReq := &team_proto.GetTeamMembershipRequest{
		UserId: []string{reqUserID.String()},
		TeamId: team.TeamID.String(),
	}

	membersResp, err := es.teamClient.CheckTeamMembership(ctx, teamMembersReq)
	if err != nil {
		es.l.Error("Error checking team membership response")
		return nil, err
	}

	userMembersMap := membersResp.Members[reqUserID.String()]

	isAuthorized := userMembersMap.Role == "coach" || userMembersMap.Role == "manager"

	if !isAuthorized {
		es.l.Error("User NOT allowed to update Event details")
		return nil, ErrForbidden
	}

	//perform database write
	//updatedEvent,err := es.UpdateEvent
	//query := `UPDATE event SET  name=$1,location=$2,start_time=$3,end_time=$4 WHERE event_id=$1`
	updatedEvent, err := es.UpdateEvent(ctx, team.ID, toUpdate.Title, toUpdate.Location, toUpdate.StartTime, toUpdate.EndTime)
	if err != nil {
		//
		return nil, err
	}

	return updatedEvent, nil
}

func (es *EventService) UpdateEvent(ctx context.Context, eventID uuid.UUID, name string, location string, start time.Time, end time.Time) (*models.Event, error) {
	es.l.Info("update team details database write")

	var updateEvent models.Event

	query := `UPDATE events SET name=$1,location=$2,start_time=$3,end_time=$4 WHERE event_id=$5
	RETURNING event_id,team_id,name,event_type,location,start_time,endtime`

	err := es.db.QueryRowContext(ctx, query, name, location, start, end).Scan(
		&updateEvent.Title,
		&updateEvent.EventType,
		&updateEvent.Location,
		&updateEvent.StartTime,
		&updateEvent.EndTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			es.l.Error("")
			return nil, ErrNotFound

		}
		return nil, fmt.Errorf("other database errors: %v", err)
	}
	return &updateEvent, nil
}

// DELETE EVENT SERVICE OPEARATION

/*
func (es *EventService) CreateBulkInsert(ctx context.Context, txs pgx.Tx, attendances []models.Attendance) error {
	es.l.Info("Successfully starting bulk insert .... ")

	//pre-allocating outer slice
	dataRows := make([][]interface{}, 0, len(attendances))

	for _, attendance := range attendances {
		dataRows = append(dataRows, []interface{}{
			attendance.EventID,
			attendance.TeamID,
			attendance.UserID,
			attendance.Status,
			attendance.UpdateteAt,
		})
	}

	//execution ->defining tbl indentifiers and colmns

	tableIdentifier := pgx.Identifier{"event_attendance"}
	columnNames := []string{"event_id", "team_id", "user_id", "status", "updatedat"}

	//perform bulk insert -> use CopyFrom method
	rowsAffected, err := txs.CopyFrom(ctx, tableIdentifier, columnNames, pgx.CopyFromRows(dataRows))
	if err != nil {
		return fmt.Errorf("failed to execute bulk insert most probably due to: %w", err)
	}

	if int(rowsAffected) != len(attendances) {
		return fmt.Errorf("bulk insert mismatch, expected to insert %d but only %d was inserted", len(attendances), rowsAffected)
	}
	return nil
}


func (es *EventService) CreateAttendanceList(ctx context.Context, tx *sql.Tx, event_id, userID uuid.UUID, teamID uuid.UUID, status string, updatedat time.Time) (*models.Attendance, error) {
	es.l.Info("Creating attendance list")

	attendance, err := models.NewAttendance(event_id, teamID, userID, status, updatedat)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO attendance(event_id, team_id,user_id, status, updatedat) VALUES($1, $2,$3,$4,$5,$6)`

	_, err = es.db.ExecContext(ctx, query, attendance.EventID, attendance.TeamID, attendance.UserID, attendance.Status, attendance.UpdateteAt)
	if err != nil {
		return nil, err
	}

	return &models.Attendance{
		EventID:    uuid.New(),
		TeamID:     teamID,
		UserID:     userID,
		Status:     status,
		UpdateteAt: time.Now(),
	}, nil
}
*/
