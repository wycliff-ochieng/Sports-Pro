package grpc

import (
	"github/wycliff-ochieng/internal/service"
	"log"

	"github.com/wycliff-ochieng/sports-proto/team_grpc/team_proto"
)

type Server struct {
	team_proto.UnimplementedTeamRPCServer
	service *service.TeamService
	logger  *log.Logger
}
