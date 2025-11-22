package api

import (
	"github/wycliff-ochieng/internal/config"
	"github/wycliff-ochieng/internal/database"
	"github/wycliff-ochieng/internal/handlers"
	internal "github/wycliff-ochieng/internal/producer"
	"github/wycliff-ochieng/internal/service"
	"net"

	rpc "github/wycliff-ochieng/grpc"

	corshandlers "github.com/gorilla/handlers"

	"log"
	"log/slog"
	"net/http"
	"os"

	auth "github.com/wycliff-ochieng/sports-common-package/middleware"
	"github.com/wycliff-ochieng/sports-common-package/team_grpc/team_proto"
	"github.com/wycliff-ochieng/sports-common-package/user_grpc/user_proto"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type APIServer struct {
	addrr string
	cfg   *config.Config
}

func NewAPIServer(addrr string, cfg *config.Config) *APIServer {
	return &APIServer{
		addrr: addrr,
		cfg:   cfg,
	}
}

func (s *APIServer) Run() {
	l := log.New(os.Stdout, ">>>TEAM SERVICE FIRING", log.LstdFlags)

	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	handler := slog.NewTextHandler(os.Stdout, handlerOpts)

	baseLogger := slog.New(handler)

	logger := baseLogger.With("service", "team-service")

	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		log.Printf("error configuring db: %v", err)
	}

	p, err := internal.InitKafkaProducer()
	if err != nil {
		log.Fatalf("something failed when initializing: %s", err)
	}

	ep := internal.NewUpdateTeam(p, "team_events")

	userServiceAddress := "localhost:50051" // "user-service-svc:50051"  -> K8s name and grpc port

	conn, err := grpc.NewClient(userServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("ERROR setting up client: %v", err)
	}

	defer conn.Close()

	userClient := user_proto.NewUserServiceRPCClient(conn)

	ts := service.NewTeamService(db, userClient, ep)

	th := handlers.NewTeamHandler(l, ts)

	//instatiate middleware
	authMiddleware := auth.AuthMiddleware(s.cfg.JWTSecret, logger) //.TeamMiddlware(s.cfg.JWTSecret)

	//set up router
	router := mux.NewRouter()

	//set up grpc server implementation
	gRPCAddress := "50052"
	lis, err := net.Listen("tcp", ":"+gRPCAddress)
	if err != nil {
		//handle this error
		log.Fatalf("ERROR spinning up network listener due to: %v", err)
	}

	grpcServ := grpc.NewServer()

	grpcServer := &rpc.Server{
		Service: ts,
		Logger:  l,
	}

	//user_proto.RegisterUserServiceRPCServer(grpcServ, grpcServer)
	team_proto.RegisterTeamRPCServer(grpcServ, grpcServer)

	//getUser := router.Methods("GET").Subrouter()
	//getUser.HandleFunc("/grab", uh.GetUserProfile)

	/*l.Printf("gRPC server listening o port: %v", gRPCAddress)
	if err := grpcServ.Serve(lis); err != nil {
		log.Fatalf("Some error spinning up RPC server: %v", err)
		os.Exit(1)
	}*/

	go func() {
		l.Printf("gRPC server starting on port: %v", gRPCAddress)
		if err := grpcServ.Serve(lis); err != nil {
			log.Fatalf("Fatal error: gRPC server failed to serve: %v", err)
		}
	}()

	//routes
	createTeam := router.Methods("POST").Subrouter()
	createTeam.HandleFunc("/api/teams", th.CreateTeam)
	createTeam.Use(authMiddleware)
	createTeam.Use(auth.RequireRole("coach", "manager", "player"))

	getTeams := router.Methods("GET").Subrouter()
	getTeams.HandleFunc("/api/get/teams", th.GetTeams)
	getTeams.Use(authMiddleware)

	getTeamsByID := router.Methods("GET").Subrouter()
	getTeamsByID.Use(authMiddleware)
	getTeamsByID.HandleFunc("/api/team/{team_id}", th.GetTeamsByID)

	updateTeam := router.Methods("PUT").Subrouter()
	updateTeam.HandleFunc("/api/team/{team_id}/update", th.UpdateTeam)
	updateTeam.Use(authMiddleware)
	updateTeam.Use(auth.RequireRole("coach", "manager", "player"))
	//updateTeam.Use(middleware.UserMiddlware(s.cfg.JWTSecret))

	addMember := router.Methods("POST").Subrouter()
	addMember.HandleFunc("/api/team/{team_idid}/add", th.AddTeamMember)
	addMember.Use(authMiddleware)
	addMember.Use(auth.RequireRole("coach", "manager", "player"))

	getTeamList := router.Methods("GET").Subrouter()
	getTeamList.HandleFunc("/api/team/{team_id}/members", th.GetTeamRoster)

	updateTeamMember := router.Methods("PUT").Subrouter()
	updateTeamMember.HandleFunc("/api/team/{teamid}/members/{user_id}/update", th.UpdateTeamMember)
	updateTeamMember.Use(auth.RequireRole("coach", "manager"))

	deleteTeamMember := router.Methods("DELETE").Subrouter()
	deleteTeamMember.HandleFunc("/api/team/{teamid}/member/{user_id}/delete", th.RemoveTeamMember)
	deleteTeamMember.Use(auth.RequireRole("coach", "manager"))

	origins := s.cfg.CORSAllowedOrigins

	allowedMethods := corshandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := corshandlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	allowCredentials := corshandlers.AllowCredentials()
	allowedOrigins := corshandlers.AllowedOrigins(origins)

	cm := corshandlers.CORS(allowedOrigins, allowCredentials, allowedMethods, allowedHeaders)(router)

	if err := http.ListenAndServe(s.addrr, cm); err != nil {
		log.Fatalf("Error setting up router: %v", err)
	}

}
