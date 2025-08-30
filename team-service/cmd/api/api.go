package api

import (
	"github/wycliff-ochieng/internal/config"
	"github/wycliff-ochieng/internal/database"
	"github/wycliff-ochieng/internal/handlers"
	internal "github/wycliff-ochieng/internal/producer"
	"github/wycliff-ochieng/internal/service"
	"github/wycliff-ochieng/middleware"
	"log"
	"net/http"
	"os"

	"github.com/wycliff-ochieng/sports-proto/user_grpc/user_proto"

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

	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		log.Printf("error configuring db: %v", err)
	}

	p, err := internal.InitKafkaProducer()
	if err != nil {
		log.Fatalf("something failed when initializing: %s", err)
	}

	ep := internal.NewUpdateTeam(p, "team_events")

	userServiceAddress := "50051" // "user-service-svc:50051"  -> K8s name and grpc port

	conn, err := grpc.NewClient(userServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("ERROR setting up client: %v", err)
	}

	defer conn.Close()

	userClient := user_proto.NewUserServiceRPCClient(conn)

	ts := service.NewTeamService(db, userClient, ep)

	th := handlers.NewTeamHandler(l, ts)

	//instatiate middleware
	authMiddleware := middleware.TeamMiddlware(s.cfg.JWTSecret)

	//set up router
	router := mux.NewRouter()

	//routes
	createTeam := router.Methods("POST").Subrouter()
	createTeam.HandleFunc("/api/teams", th.CreateTeam)
	createTeam.Use(authMiddleware)
	createTeam.Use(middleware.RequireRole("coach", "manager"))

	getTeams := router.Methods("GET").Subrouter()
	getTeams.HandleFunc("/api/get/teams", th.GetTeams)

	getTeamsByID := router.Methods("GET").Subrouter()
	getTeamsByID.HandleFunc("/api/team/{team_id}", th.GetTeamsByID)

	updateTeam := router.Methods("PUT").Subrouter()
	updateTeam.HandleFunc("/api/team/{teamid}/update", th.UpdateTeam)
	updateTeam.Use(middleware.RequireRole("COACH", "MANAGER"))
	//updateTeam.Use(middleware.UserMiddlware(s.cfg.JWTSecret))

	addMember := router.Methods("POST").Subrouter()
	addMember.HandleFunc("/api/team/{teamid}/add", th.AddTeamMember)
	addMember.Use(middleware.RequireRole("coach", "manager"))

	getTeamList := router.Methods("GET").Subrouter()
	getTeamList.HandleFunc("/api/team/{teamid}/members", th.GetTeamRoster)

	updateTeamMember := router.Methods("PUT").Subrouter()
	updateTeamMember.HandleFunc("/api/team/{teamid}/members/{userid}/update", th.UpdateTeamMember)
	updateTeamMember.Use(middleware.RequireRole("coach", "manager"))

	deleteTeamMember := router.Methods("DELETE").Subrouter()
	deleteTeamMember.HandleFunc("/api/team/{teamid}/member/{userid}/delete", th.RemoveTeamMember)
	deleteTeamMember.Use(middleware.RequireRole("coach", "manager"))

	if err := http.ListenAndServe(s.addrr, router); err != nil {
		log.Fatalf("Error setting up router: %v", err)
	}

}
