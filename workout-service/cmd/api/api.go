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
	"github.com/wycliff-ochieng/sports-common-package/user_grpc/user_proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	addr string
	cfg  *config.Config
}

func NewAPIServer(addr string, cfg *config.Config) *Server {
	return &Server{
		addr: addr,
		cfg:  cfg,
	}
}

func (s *Server) Run() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	//setup middleware
	authMiddleware := auth.AuthMiddleware(s.cfg.JWTSecret, logger)

	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		log.Printf("setting up database error: %s", err)
	}

	userServiceAddress := "localhost:50051" // "user-service-svc:50051"  -> K8s name and grpc port

	conn, err := grpc.NewClient(userServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("ERROR setting up client: %v", err)
	}

	defer conn.Close()

	userClient := user_proto.NewUserServiceRPCClient(conn)

	ws := service.NewWorkoutService(db, userClient)

	wh := handlers.NewWorkoutHandler(logger, ws)

	router := mux.NewRouter()

	createWorkout := router.Methods("POST").Subrouter()
	createWorkout.HandleFunc("/api", wh.CreateWorkout)
	createWorkout.Use(authMiddleware)

	getWorkouts := router.Methods("GET").Subrouter()
	getWorkouts.HandleFunc("/api/workout", wh.GetAllWorkouts)

	if err := http.ListenAndServe(s.addr, router); err != nil {
		log.Printf("error listening to the address, %s", err)
	}
}
