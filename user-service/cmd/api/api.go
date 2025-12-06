package api

import (
	"context"
	"log"
	"net"

	//"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	corshandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	rpc "github.com/wycliff-ochieng/grpc"
	"github.com/wycliff-ochieng/internal/config"
	"github.com/wycliff-ochieng/internal/consumer"
	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/handlers"
	internal "github.com/wycliff-ochieng/internal/producer"
	"github.com/wycliff-ochieng/internal/service"
	appmiddleware "github.com/wycliff-ochieng/middleware"
	"github.com/wycliff-ochieng/sports-common-package/user_grpc/user_proto"
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
	//bootstrapServers := "localhost:9092"

	bootstrapServers := os.Getenv("KAFKA_BROKER")

	if bootstrapServers == "" {
		bootstrapServers = "localhost:9092"
	}
	groupID := "foo"
	topic := "profiles"

	jwtSecret := s.cfg.JWTSecret
	if jwtSecret == "" {
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
	router.Use(appmiddleware.UserMiddlware(jwtSecret))

	getUserProfile := router.Methods("GET").Subrouter()
	getUserProfile.HandleFunc("/profile/get", uh.GetProfileByUUID)

	updateUser := router.Methods("PUT").Subrouter()
	updateUser.HandleFunc("/update", uh.UpdateUserProfile)

	//gRPC server configuration
	gRPCAddress := "50051"
	lis, err := net.Listen("tcp", ":"+gRPCAddress)
	if err != nil {
		//handle this error
		log.Fatalf("ERROR spinning up network listener due to: %v", err)
	}

	grpcServ := grpc.NewServer()

	grpcServer := &rpc.Server{
		Service: us,
		Logger:  l,
	}

	user_proto.RegisterUserServiceRPCServer(grpcServ, grpcServer)

	//getUser := router.Methods("GET").Subrouter()
	//getUser.HandleFunc("/grab", uh.GetUserProfile)

	// Start gRPC server in a goroutine so HTTP server can also start
	l.Printf("gRPC server listening o port: %v", gRPCAddress)
	go func() {
		if err := grpcServ.Serve(lis); err != nil {
			log.Fatalf("Some error spinning up RPC server: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		grpcServ.GracefulStop()
	}()

	origins := s.cfg.CORSAllowedOrigins

	allowedMethods := corshandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := corshandlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	allowCredentials := corshandlers.AllowCredentials()
	allowedOrigins := corshandlers.AllowedOrigins(origins)

	cm := corshandlers.CORS(allowedOrigins, allowCredentials, allowedMethods, allowedHeaders)(router)

	if err := http.ListenAndServe(s.addr, cm); err != nil {
		log.Printf("Error listeniing %v", err)
	}
}
