package api

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/wycliff-ochieng/internal/config"
	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/handlers"
	"github.com/wycliff-ochieng/internal/service"
	"github.com/wycliff-ochieng/sports-proto/team_grpc/team_proto"
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

	l := log.New(os.Stdout, "EVENT SERVICE UP AND RUNNING", log.LstdFlags)

	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		log.Fatalf("error setting up postgres connection due to: %v", err)
	}

	teamServiceAddress := "50051"

	conn, err := grpc.NewClient(teamServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error setting up grpc client: %v", err)
	}

	defer conn.Close()

	//set up teamServiceClient
	teamClient := team_proto.NewTeamRPCClient(conn)

	es := service.NewEventService(db, teamClient)

	eh := handlers.NewEventHandler(l, es)

	router := mux.NewRouter()

	createEvent := router.Methods("POST").Subrouter()
	createEvent.HandleFunc("/api/events/new", eh.CreateEvent)

	if err := http.ListenAndServe(s.addr, router); err != nil {
		log.Fatalf("issue with pinging router")
	}
}
