package grpc

import (
	"context"
	"github/wycliff-ochieng/internal/service"
	"log"

	"github.com/google/uuid"
	"github.com/wycliff-ochieng/sports-common-package/team_grpc/team_proto"
)

type Server struct {
	team_proto.UnimplementedTeamRPCServer
	Service *service.TeamService
	Logger  *log.Logger
}

func NewServer(service *service.TeamService, l *log.Logger) *Server {
	return &Server{
		Service: service,
		Logger:  l,
	}
}

func (s *Server) CheckTeamMembership(ctx context.Context, req *team_proto.GetTeamMembershipRequest) (*team_proto.GetTeamMembershipResponse, error) {

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		log.Printf("error: %s", err)
		return nil, err
	}

	//req.TeamId

	members, err := s.Service.GetTeamsMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	grpcTeamMembers := make(map[string]*team_proto.TeamMember)

	for _, m := range members {
		grpcTeamMembers[m.UserID.String()] = &team_proto.TeamMember{
			UserId: m.UserID.String(),
			TeamId: m.TeamID.String(),
			Role:   m.Role,
		}

	}

	return &team_proto.GetTeamMembershipResponse{Members: grpcTeamMembers}, nil

}

func (s *Server) GetTeamSummary(ctx context.Context, req *team_proto.GetTeamSummaryRequest) (*team_proto.GetTeamSummaryResponse, error) {

	teamID, err := uuid.Parse(req.TeamId)
	if err != nil {
		return nil, err
	}

	members, err := s.Service.GetTeamsMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}

	//grpcTeamMembers := make(map[string]*team_proto.TeamMember)
	var grpcTeamMembers []*team_proto.TeamMember
	for _, m := range members {
		grpcTeamMembers = append(grpcTeamMembers, &team_proto.TeamMember{
			UserId: m.UserID.String(),
			TeamId: m.TeamID.String(),
			Role:   m.Role,
		})
	}
	return &team_proto.GetTeamSummaryResponse{Members: grpcTeamMembers}, nil

}
