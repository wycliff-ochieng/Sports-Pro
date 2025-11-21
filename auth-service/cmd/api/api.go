package cmd

import (
	"log"
	"net/http"
	"os"
	"sports/authservice/internal/config"
	"sports/authservice/internal/database"
	"sports/authservice/internal/handlers"
	internal "sports/authservice/internal/producer"
	"sports/authservice/internal/service"

	corshandlers "github.com/gorilla/handlers"

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
	ep := internal.NewCreateUser(p, "profiles")

	go ep.DeliveryReportHandler()

	ah := handlers.NewAuthHandler(l, sh, ep)

	registerRouter := router.Methods("POST").Subrouter()
	registerRouter.HandleFunc("/register", ah.Register)

	loginRouter := router.Methods("POST").Subrouter()
	loginRouter.HandleFunc("/login", ah.Login)

	//CORS configuration

	allowedMethods := corshandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := corshandlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	allowCredentials := corshandlers.AllowCredentials()

	cm := corshandlers.CORS(allowCredentials, allowedMethods, allowedHeaders)(router)

	if err := http.ListenAndServe(s.addr, cm); err != nil {
		log.Fatalf("error listening to server:%v", err)
	}

}
