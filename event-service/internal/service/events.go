package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/sports-proto/team_grpc/team_proto"
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
	l          *slog.Logger
}

func NewEventService(db database.DBInterface, teamClient team_proto.TeamRPCClient) *EventService {
	return &EventService{
		db:         db,
		teamClient: teamClient,
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
