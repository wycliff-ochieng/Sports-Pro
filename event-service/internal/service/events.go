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
	CreateTeamEvent(ctx context.Context, eventID uuid.UUID, teamID uuid.UUID, eventType string, startTime time.Time, endTime time.Time) (*models.Event, error)
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

func NewEventService(db database.DBInterface, teamClient team_proto.TeamRPCClient, userCllient user_proto.UserServiceRPCClient) *EventService {
	return &EventService{
		db:         db,
		teamClient: teamClient,
		userClient: userCllient,
	}
}

func (es *EventService) CreateTeamEvent(ctx context.Context, reqUserID uuid.UUID, eventID uuid.UUID, teamID uuid.UUID, eventType string, startTime time.Time, endTime time.Time) (*models.Event, error) {
	//Authorization via gRPC
	es.l.Info("Creation of event initiated by user")

	var CreateEventReq models.CreateEventReq

	membershipReq := &team_proto.GetTeamMembershipRequest{
		TeamId: CreateEventReq.TeamID.String(),
		UserId: []string{reqUserID.String()},
	}

	members, err := es.teamClient.CheckTeamMembership(ctx, membershipReq)
	if err != nil {
		return nil, err
	}

	//check response/error from grRPC call- > if user is a member of that team or not and if member is coach/manager
	membersInfo, found := members.Members[reqUserID.String()]
	if !found {
		es.l.Error("user does not exist in the team therefore is not a member")
		return nil, ErrForbidden
	}

	isAuthorized := membersInfo.Role == "coach" || membersInfo.Role == "manager"
	if !isAuthorized {
		es.l.Warn("user is not a coach/manager therefore cant create events")
	}
	es.l.Info("Authorization is successfull")

	//get team data for attendance table insert(business logic)
	teamMembersReq := team_proto.GetTeamSummaryRequest{TeamId: CreateEventReq.TeamID.String()}
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

	newEvent := &models.Event{
		TeamID:    CreateEventReq.TeamID,
		Title:     CreateEventReq.Name,
		EventType: CreateEventReq.EventType,
		Location:  CreateEventReq.Location,
		StartTime: CreateEventReq.StartTime,
		EndTime:   CreateEventReq.EndTime,
	}

	createdEvent, err := es.CreateEvent(ctx, txs, newEvent.TeamID, newEvent.Title, newEvent.EventType, newEvent.Location, newEvent.StartTime, newEvent.EndTime)
	if err != nil {
		es.l.Error("error creating event due to ", "error", err)
	}
	es.l.Info("event created successfully for team", "teamID", createdEvent.TeamID)

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
				TeamID:     createdEvent.TeamID,
				Status:     "PENDING",
				UpdateteAt: time.Now(),
			})
		}

		//insert into attendance table (bulk insert->provides high performance)
		if err := es.CreateAttendanceInsert(ctx, txs, attendanceRecords); err != nil {
			es.l.Error("error inserting to iniital attendance table")
		}
		es.l.Info("bulk rcord attendance created ")

		if err := txs.Commit(); err != nil {
			log.Fatalf("error commiting /creating team event due to: %v", err)
			return nil, err
		}
		es.l.Info("successfully created event for team %s", "teamID", teamID)

	}

	return nil, nil
}

func (es *EventService) CreateEvent(ctx context.Context, tx *sql.Tx, teamID uuid.UUID, name string, eventtype string, location string, starttime, endtime time.Time) (*models.Event, error) {
	es.l.Info("Create event database execution")

	//var event *models.Event

	newEvent, err := models.NewEvent(teamID, name, eventtype, location, starttime, endtime)
	if err != nil {
		es.l.Error("error creating new team event")
		return nil, fmt.Errorf("due to :%v", err)
	}

	query := `INSERT INTO event(event_id,team_id,title,event_type,start_time,end_time) VALUES($1,$2,$3,$4,$5,$6)`

	_, err = es.db.ExecContext(ctx, query, newEvent.TeamID, newEvent.Title, newEvent.EventType, newEvent.Location, newEvent.StartTime, newEvent.EventType)
	if err != nil {
		log.Fatalf("Error executing query:%v", err)
		return nil, err
	}
	return &models.Event{
		ID:        uuid.New(),
		TeamID:    teamID,
		Title:     name,
		EventType: eventtype,
		Location:  location,
		StartTime: starttime,
		EndTime:   endtime,
	}, nil
}

func (es *EventService) CreateAttendanceInsert(ctx context.Context, tx *sql.Tx, records []models.Attendance) error {
	es.l.Info("Starting bulk insert using database/sql driver")

	if len(records) == 0 {
		return nil
	}

	sqlStr := `INSERT INTO attendance VALUES(event_id,user_id,status)`

	vals := []interface{}{}

	for i, record := range records {
		placeHolder1 := i*3 + 1
		placeHolder2 := i*3 + 2
		placeHolder3 := i*3 + 3

		sqlStr += fmt.Sprintf("(%d,%d,%d),", placeHolder1, placeHolder2, placeHolder3)

		vals = append(vals, record.EventID, record.UserID, record.Status)
	}

	sqlStr = strings.TrimSuffix(sqlStr, ",")

	stmt, err := tx.PrepareContext(ctx, sqlStr)
	if err != nil {
		return fmt.Errorf("failed to prepare bulk insert due to: %v", err)
	}

	result, err := stmt.ExecContext(ctx, vals)
	if err != nil {
		return fmt.Errorf("failed to execute bulk insert: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("faileed to get rows affected, %v", err)
	}

	if int(rowsAffected) != len(records) {
		return fmt.Errorf("bulk insert mismatch %v", err)
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

	//gRPC call to team service to check membership
	memberShipReq := &team_proto.GetTeamMembershipRequest{
		TeamId: teamID,
		UserId: []string{reqUserID.String()},
	}

	members, err := es.teamClient.CheckTeamMembership(ctx, memberShipReq)
	if err != nil {
		return nil, err
	}

	membersInfo, found := members.Members[teamID]
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

	//call get events
	event, err := es.GetEvent(ctx, eventID)
	if err != nil {
		es.l.Error("Error getting events from database to repository layer")
		log.Fatalf("The error: %v", err)
		return nil, err
	}

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

	//gRPC data enrichment
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

	query := `SELECT user_id,status FROM attendance WHERE event_id=$1`

	rows, err := es.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var attendance models.Attendance
		err = rows.Scan(
			&attendance.EventID,
			&attendance.TeamID,
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

	query := `SELECT event_id,name,event_type,location,start_time,end_time FROM events WHERE event_id = $1`

	err := es.db.QueryRowContext(ctx, query, eventID).Scan(
		&event.ID,
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

func (es *EventService) UpdateEventDetails(ctx context.Context, reqUserID uuid.UUID, teamID uuid.UUID, eventID uuid.UUID, toUpdate models.UpdateEventReq) (*models.EventDetails, error) {
	es.l.Info("PUT operation for the event service")

	txs, err := es.db.BeginTx(ctx, nil)
	if err != nil {
		es.l.Error("Error while trying to begin transactions")
	}

	defer txs.Rollback()

	return nil, nil
}

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
