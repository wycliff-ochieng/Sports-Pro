package api

import (
	"log"

	"github.com/wycliff-ochieng/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type APIServer struct {
	addr string
	cfg  *config.Config
}

func NewAPIServer(addr string, cfg *config.Config) *APIServer {
	return &APIServer{
		addr: addr,
		cfg:  cfg,
	}
}

func (s *APIServer) Run() {

	teamServiceAddress := "50051"

	conn, err := grpc.NewClient(teamServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error setting up grpc client: %v", err)
	}

	defer conn.Close()

	//set up teamServiceClient
}
