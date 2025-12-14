package api

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	corshandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/wycliff-ochieng/internal/config"
	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/filestore"
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
	bucketName := s.cfg.MinIOBucket   //os.Getenv("MINIO_BUCKET")    //"sportspro"
	useSSL := false                   //os.Getenv("MINIO_USE_SSL")
	accesskey := s.cfg.MinIOAccessKey //os.Getenv("MINIO_ACCESS_KEY") //"4XET6XMT3LP810RYWGCO"
	secretKey := s.cfg.MinIOSecretKey //os.Getenv("MINIO_SECRET_KEY") //"2ai9tXU0mGV+1gVxQeEAfhSv+SgbOMKekRE6PqOA"
	endpoint := s.cfg.MinIOEndpoint   //"localhost:3000"               //os.Getenv("MINIO_ENDPOINT")    //localhost:9001

	//setup middleware
	authMiddleware := auth.AuthMiddleware(s.cfg.JWTSecret, logger)

	db, err := database.NewPostgresDB(s.cfg)
	if err != nil {
		log.Printf("setting up database error: %s", err)
	}

	fs, err := filestore.NewFileStore(endpoint, accesskey, secretKey, bucketName, useSSL)
	if err != nil {
		log.Printf("issue setting up minio due to : %s", err)
	}

	//userServiceAddress := "localhost:50051" // "user-service-svc:50051"  -> K8s name and grpc portuserServiceAddress := os.Getenv("USER_SERVICE_GRPC_ADDR")
	userServiceAddress := os.Getenv("USER_SERVICE_GRPC_ADDR")
	if userServiceAddress == "" {
		userServiceAddress = "localhost:50051"
	}

	conn, err := grpc.NewClient(userServiceAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("ERROR setting up client: %v", err)
	}

	defer conn.Close()

	userClient := user_proto.NewUserServiceRPCClient(conn)

	ws := service.NewWorkoutService(db, userClient, fs)

	wh := handlers.NewWorkoutHandler(logger, ws)

	router := mux.NewRouter()

	createWorkout := router.Methods("POST").Subrouter()
	createWorkout.HandleFunc("/api", wh.CreateWorkout)
	createWorkout.Use(authMiddleware)

	getWorkouts := router.Methods("GET").Subrouter()
	getWorkouts.HandleFunc("/api/workout", wh.GetAllWorkouts)
	getWorkouts.Use(authMiddleware)

	createExercise := router.Methods("POST").Subrouter()
	createExercise.HandleFunc("/api/exercise", wh.CreateExercise)
	createExercise.Use(authMiddleware)

	getExercises := router.Methods("GET").Subrouter()
	getExercises.HandleFunc("/api/exercise/fetch", wh.GetAllExercises)
	getExercises.Use(authMiddleware)

	preURL := router.Methods("POST").Subrouter()
	preURL.HandleFunc("/api/media/presigned-url", wh.MediaPresignedURL)
	preURL.HandleFunc("/api/media/upload-complete", wh.UploadComplete)
	preURL.Use(authMiddleware)

	origins := s.cfg.CORSAllowedOrigins

	allowedMethods := corshandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := corshandlers.AllowedHeaders([]string{"Content-Type", "Authorization"})
	allowCredentials := corshandlers.AllowCredentials()
	allowedOrigins := corshandlers.AllowedOrigins(origins)

	cm := corshandlers.CORS(allowedOrigins, allowCredentials, allowedMethods, allowedHeaders)(router)

	if err := http.ListenAndServe(s.addr, cm); err != nil {
		log.Printf("error listening to the address, %s", err)
	}
}
