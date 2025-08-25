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
	profiles, err := s.service.GetUserProfilesByUUIDs(ctx, req.Userid)
	if err != nil {
		s.Logger.Error("Issue getting user profiles from user-service")
		return nil, err
	}

	//convert Profiles struct to gRPC userProfile struct
	grpcProfile := make(map[string]*grpc.UserProfile)
	for _, p := range profiles {
		grpcProfile[p.UserID.String()] = &grpc.UserProfile{
			Userid:    p.UserID.String(),
			Firstname: p.Firstname,
			Lastname:  p.Lastname,
			Email:     p.Email,
		}
	}

	return nil, nil
}
