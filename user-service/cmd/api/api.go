package api

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wycliff-ochieng/internal/config"
	"github.com/wycliff-ochieng/internal/consumer"
	"github.com/wycliff-ochieng/internal/database"
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

	//gracefully shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	//db setup
	db, err := database.Newpostgres(s.cfg)
	if err != nil {
		log.Fatalf("something fatal hapenned: %v", err)
	}

	//set up repo service
	us := service.NewUserService(db)

	//set up kafka
	ks, err := consumer.NewUserEventConsumer(l, us, bootstrapServers, groupID)
	if err != nil {
		log.Fatalf("error setting up consumer: %v", err)
	}

	//set up consumer to start background in a background goroutine
	go ks.StartEventConsumer(ctx, topic)

	//set up router
}
