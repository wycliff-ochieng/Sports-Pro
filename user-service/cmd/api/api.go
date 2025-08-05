package api

import (
	"context"
	"log"
	//"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/wycliff-ochieng/internal/config"
	"github.com/wycliff-ochieng/internal/consumer"
	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/handlers"
	internal "github.com/wycliff-ochieng/internal/producer"
	"github.com/wycliff-ochieng/internal/service"
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
	l := log.New(os.Stdout, ">>USER_SERVICE FIRING", log.LstdFlags)
	bootstrapServers := "localhost:9092"
	groupID := "foo"
	topic := "profiles"

	jwtSecret := s.cfg.JWTSecret
	if jwtSecret == ""{
		log.Printf("no secret found")
	}

	//configure middleware instance

	//protect routes

	//gracefully shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	//db setup
	db, err := database.Newpostgres(s.cfg)
	if err != nil {
		log.Fatalf("something fatal hapenned: %v", err)
	}
	//set up kafka producer
	p, err := internal.InitKafkaProducer()
	if err != nil {
		log.Fatalf("something failed when initializing: %s", err)
	}

	ep := internal.NewUpdateUser(p, "profile")

	//set up repo service
	us := service.NewUserService(l, db, ep)

	//set up kafka consumer
	ks, err := consumer.NewUserEventConsumer(l, us, bootstrapServers, groupID)
	if err != nil {
		log.Fatalf("error setting up consumer: %v", err)
	}

	//set up consumer to start background in a background goroutine
	go ks.StartEventConsumer(ctx, topic)

	//set up router
	uh := handlers.NewUserHandler(l, us)

	router := mux.NewRouter()

	updateUser := router.Methods("PUT").Subrouter()
	updateUser.HandleFunc("/update", uh.UpdateUserProfile)

	//getUser := router.Methods("GET").Subrouter()
	//getUser.HandleFunc("/grab", uh.GetUserProfile)

	if err := http.ListenAndServe(s.addr, router); err != nil {
		log.Printf("Error listeniing %v", err)
	}
}
