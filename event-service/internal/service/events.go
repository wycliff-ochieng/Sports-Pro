package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/sports-proto/team_grpc/team_proto"
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
	l *slog.Logger
}

func NewEventService(db database.DBInterface, teamClient team_proto.TeamRPCClient) *EventService {
	return &EventService{
		db:         db,
		teamClient: teamClient,
	}
}

func (es *EventService) CreateTeamEvent(ctx context.Context,reqUserID uuid.UUID,eventID uuid.UUID, teamID uuid.UUID, eventType string, startTime time.Time, endTime time.Time) (*models.Event, error) {
	//Authorization via gRPC
	es.l.Info("Creation of event initiated by user")

	var CreateEventReq models.CreateEventReq
	
	membershipReq := &team_proto.GetTeamMembershipRequest{
		TeamId: CreateEventReq.TeamID.String(),
		UserId:[]string{reqUserID.String()},
	}

	members, err := es.teamClient.CheckTeamMembership(ctx,membershipReq)
	if err != nil{
		return nil,err
	}
	return nil, nil
}
