package service

import (
	//"strings"

	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (s *WorkoutService) CreateWorkout(ctx context.Context, reqUserID uuid.UUID, name, category, description string, exerc []models.WorkoutExerciseResponse /*workout models.Workout, exerc models.Exercise /*workout handlers.CreateWorkoutReq*/) (*models.WorkoutExerciseResponse, error) {

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
	//excercisIDs := []uuid.UUIDs{}
	var wkout models.Exercise

	exerciseID, err := s.CollectExerciseID(&wkout)
	if err != nil {
		//handle err
	}

	err = s.ValidateExerciseID(ctx, exerciseID)
	if err != nil {
		//handle err
	}

	//begin transaction
	txs, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transactions: %s", err)
	}

	defer txs.Rollback()

	//newWorkoutID := wkout.ID

	//exerciseQuery := `INSERT INTO workout_exercise() VALUES($1,$2,$3,$4)`

	//err = txs.QueryRowContext(ctx, exerciseQuery).Scan()

	wkt, err := s.CreateWorkoutRepo(ctx, txs, name, category, description, reqUserID)
	if err != nil {
		//handle error
	}

	if err := s.CreateExecrciseRepo(ctx, txs, exerc); err != nil {
		log.Printf("iss")
	}
	//if err !=

	return &models.WorkoutExerciseResponse{
		WorkoutID:   wkt.ID,
		Name:        wkt.Name,
		ExcerciseID: wkout.ID,
		//Order: ,
	}, nil
}

func (s *WorkoutService) CreateWorkoutRepo(ctx context.Context, tx *sql.Tx, name string, category string, description string, createdby uuid.UUID) (*models.Workout, error) {

	var createdWorkout models.Workout

	query := `INSERT INTO workout(name,category,description,created_by)  VALUES($1,$2,$3) RETERNING id`

	err := s.db.QueryRowContext(ctx, query, name, category, description, createdby).Scan(
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

func (s *WorkoutService) CreateExecrciseRepo(ctx context.Context, tx *sql.Tx, exercises []models.WorkoutExerciseResponse) error {

	//bulk insert

	if len(exercises) == 0 {
		return nil
	}

	sqlStr := `INSERT INTO workout_exercise(id,name,description,instructions,createdby) VALUES($1,$2,$3,$4,$5)`

	vals := []interface{}{}

	for i, exercise := range exercises {
		placeHolder1 := i*5 + 1
		placeHolder2 := i*5 + 2
		placeHolder3 := i*5 + 3
		placeHolder4 := i*5 + 4
		placeHolder5 := i*5 + 5

		placeholder := fmt.Sprintf("(%d,%d,%d,%d,%d)", placeHolder1, placeHolder2, placeHolder3, placeHolder4, placeHolder5)

		sqlStr += placeholder

		vals = append(vals, exercise.WorkoutID, exercise.ExcerciseID, exercise.Order, exercise.Sets, exercise.Reps)
	}

	sqlStr = strings.TrimSuffix(sqlStr, ",")

	stmt, err := tx.PrepareContext(ctx, sqlStr) //method that creates a prepared context for later execution
	if err != nil {
		log.Printf("handle errors from the %s", err)
		return err
	}

	result, err := stmt.ExecContext(ctx, vals)
	if err != nil {
		//handle errors
		log.Printf("issues getting results: %s", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		//handle errors
		log.Printf("error in this Ops: %s", err)
	}

	if int(rowsAffected) != len(exercises) {
		return fmt.Errorf("bulk insert mismatch, %s", err)
	}
	return nil
}

func (ws *WorkoutService) CollectExerciseID(*models.Exercise) ([]string, error) {
	var exercises []models.Exercise

	exerciseeUUIDs := make(map[string]struct{})
	IDs := make([]string, 0, len(exerciseeUUIDs))
	for _, exc := range exercises {
		excUUID := exc.ID

		//IDs := make([]string,0,len(exerciseeUUIDs))

		IDs = append(IDs, excUUID.String())
	}

	return IDs, nil
}

func (ws *WorkoutService) ValidateExerciseID(ctx context.Context, exerciseID []string) error {

	//exerciseUUID, err := uuid.Parse(exerciseID)

	var exerciseUUID []uuid.UUID
	for _, ID := range exerciseID {
		uuid, err := uuid.Parse(ID)
		if err != nil {
			//handle err
		}

		exerciseUUID = append(exerciseUUID, uuid)
	}

	query := `SELECT COUNT(exerciseID) FROM exercise WHERE exerciseID = ANY($1::exerciseID[])`

	err := ws.db.QueryRowContext(ctx, query, exerciseID) //.Scan(&)
	if err != nil {
		//handler error
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
