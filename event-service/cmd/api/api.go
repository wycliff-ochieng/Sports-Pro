package api

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/wycliff-ochieng/internal/config"
	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/handlers"
	"github.com/wycliff-ochieng/internal/service"
	auth "github.com/wycliff-ochieng/sports-common-package/middleware"
	"github.com/wycliff-ochieng/sports-common-package/team_grpc/team_proto"
	"github.com/wycliff-ochieng/sports-common-package/user_grpc/user_proto"
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

	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, handlerOpts)

	baseLogger := slog.New(handler)

	logger := baseLogger.With("service", "event-service")

	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		log.Fatalf("error setting up postgres connection due to: %v", err)
	}

	teamServiceAddress := "localhost:50052" //k8s service name and port

	conn, err := grpc.NewClient(teamServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error setting up grpc client: %v", err)
	}

	defer conn.Close()

	//set up teamServiceClient
	teamClient := team_proto.NewTeamRPCClient(conn)

	//user Client
	userClient := user_proto.NewUserServiceRPCClient(conn)

	es := service.NewEventService(db, teamClient, userClient, logger)

	eh := handlers.NewEventHandler(l, es)

	router := mux.NewRouter()

	authMiddleware := auth.AuthMiddleware(s.cfg.JWTSecret, logger)

	createEvent := router.Methods("POST").Subrouter()
	createEvent.HandleFunc("/api/events/new", eh.CreateEvent)
	createEvent.Use(authMiddleware)

	if err := http.ListenAndServe(s.addr, router); err != nil {
		log.Fatalf("issue with pinging router")
	}
}
