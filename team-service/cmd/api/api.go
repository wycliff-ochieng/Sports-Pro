package api

import (
	"github/wycliff-ochieng/internal/config"
	"github/wycliff-ochieng/internal/database"
	"github/wycliff-ochieng/internal/handlers"
	"github/wycliff-ochieng/internal/service"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

	ts := service.NewTeamService(db)

	th := handlers.NewTeamHandler(l, ts)

	//set up router
	router := mux.NewRouter()

	//routes
	createTeam := router.Methods("POST").Subrouter()
	createTeam.HandleFunc("/api/teams", th.CreateTeam)

	if err := http.ListenAndServe(s.addrr, router); err != nil {
		log.Fatalf("Error setting up router: %v", err)
	}

}
