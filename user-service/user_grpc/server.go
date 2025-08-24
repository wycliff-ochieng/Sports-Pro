package grpc

import (
	"context"
	"log/slog"

	"github.com/wycliff-ochieng/internal/service"
	grpc "github.com/wycliff-ochieng/user_grpc/user_proto"
)

type Server struct {
	grpc.GetUserRequest //forward compatibility
	service             service.UserService
	Logger              slog.Logger
}

func (s *Server) GetUserProfiles(ctx context.Context, req *grpc.GetUserRequest) (*grpc.GetUserProfileResponse, error) {
	s.Logger.Info("get profile for a user")
	//call user service for profile list
	return nil, nil
}
