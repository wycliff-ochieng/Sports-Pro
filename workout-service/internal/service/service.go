package service

import (
	//"strings"

	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/wycliff-ochieng/internal/database"

	//"github.com/wycliff-ochieng/internal/handlers"
	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/sports-common-package/user_grpc/user_proto"
)

var (
	ErrForbidden = errors.New("Not allowed to do this")
)

type WorkoutService struct {
	db         database.DBInterface
	userClient user_proto.UserServiceRPCClient
}

func NewWorkoutService(db database.DBInterface, client user_proto.UserServiceRPCClient) *WorkoutService {
	return &WorkoutService{
		db:         db,
		userClient: client,
	}
}

func (s *WorkoutService) CreateWorkout(ctx context.Context, reqUserID uuid.UUID, workout models.Workout, exerc models.Exercise /*workout handlers.CreateWorkoutReq*/) (*models.WorkoutExerciseResponse, error) {

	profilesReq := user_proto.GetUserRequest{
		Userid: strings.Split(reqUserID.String(), ""), //strings.Split(reqUserID.String(), ""),
	}

	profileRes, err := s.userClient.GetUserProfile(ctx, &profilesReq)
	if err != nil {
		log.Printf("error getting profiles response from user service : %s", err)
		return nil, err
	}

	profileMap := profileRes.Profiles

	//check if reqUserID has a profile, check also role

	profile, found := profileMap[reqUserID.String()]
	if !found {
		log.Printf("no profile for user with id : %s", reqUserID)
	} else {
		log.Printf(profile.Firstname, "%s is in the system")
	}

	//check the uuids from the exercise
	excercisIDs := []uuid.UUIDs{}

	//begin transaction
	txs, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transactions: %s", err)
	}

	//var wkout models.Workout

	//query := `INSERT INTO workout(name,category,description,created_by)  VALUES($1,$2,$3) RETERNING id`

	//err = txs.QueryRowContext(ctx, query, workout).Scan(
	//	&wkout.Name,
	//	&wkout.Category,
	//	&wkout.Description,
	//)
	if err != nil {
		return nil, err
	}

	//newWorkoutID := wkout.ID

	exerciseQuery := `INSERT INTO workout_exercise() VALUES($1,$2,$3,$4)`

	err = txs.QueryRowContext(ctx, exerciseQuery).Scan()

	//workout, err := s.CreateWorkoutRepo()

	defer txs.Rollback()

	return nil, nil
}

func (s *WorkoutService) CreateWorkoutRepo(ctx context.Context, tx *sql.Tx, workout models.Workout) (*models.Workout, error) {

	var createdWorkout models.Workout

	query := `INSERT INTO workout(name,category,description,created_by)  VALUES($1,$2,$3) RETERNING id`

	err := s.db.QueryRowContext(ctx, query, workout).Scan(
		&createdWorkout.Name,
		&createdWorkout.Category,
		&createdWorkout.Description,
		&createdWorkout.CreatedBy,
	)
	if err != nil {
		log.Printf("issue with executing inserting into workout due to: %s", err)
	}

	return &createdWorkout, nil
}

func (s *WorkoutService) CreateExecrciseRepo(ctx context.Context, tx *sql.Tx, exercises []models.Exercise) error {

	//bulk insert

	if len(exercises) == 0 {
		return nil
	}

	queryStr := `INSERT INTO workout_exercise`

	vals := []interface{}{}

	for _, exercise := range exercises {
		
	}
	return nil
}

/*
func (s *WorkoutService) CreateWorkoutRepo(ctx context.Context, ID uuid.UUID, name string, description string, caregory string, owner string, createdat time.Time, updatedat time.Time) (*models.Workout, error) {

	var wkt models.Workout

	query := `INSERT INTO workout(id,name,description,category,owner,createdat,updatedat) VALUES($1,$2,$3,$4,$5,$6,$7) RETURN id`


	query := `INSERT INTO exercise() VALUES()`

	_, err := s.db.ExecContext(ctx, query, wkt.ID, wkt.Name, wkt.Description)
	if err != nil {
		log.Printf("issue with execution: %s", err)
		return nil, err
	}
	return &wkt, nil
}
*/
