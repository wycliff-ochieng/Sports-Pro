package cmd

import (
	"log"
	"net/http"
	"os"
	"sports/authservice/internal/config"
	"sports/authservice/internal/database"
	"sports/authservice/internal/handlers"
	"sports/authservice/internal/producer"
	"sports/authservice/internal/service"

	"github.com/gorilla/mux"
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
	l := log.New(os.Stdout, "AUTHENTICATION SEVRICE API", log.LstdFlags)

	router := mux.NewRouter()

	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		log.Fatalf("something critical broke: %v", err)
	}

	p, err := internal.InitKafkaProducer()
	if err != nil {
		log.Fatalf("something failed when initializing: %s", err)
	}

	sh := service.NewAuthService(db)

	//kp := producer.PublishUserCreation()
	ep := internal.NewCreateUser(p,"profiles")

	ah := handlers.NewAuthHandler(l, sh, ep)

	registerRouter := router.Methods("POST").Subrouter()
	registerRouter.HandleFunc("/register", ah.Register)

	loginRouter := router.Methods("POST").Subrouter()
	loginRouter.HandleFunc("/login", ah.Login)

	if err := http.ListenAndServe(s.addr, router); err != nil {
		log.Fatalf("error listening to server:%v", err)
	}

}
