package service

import (
	//"strings"

	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

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

func (s *WorkoutService) CreateWorkout(ctx context.Context, reqUserID uuid.UUID, name, category, description string, exerc []models.Exercise) (*models.CreateWorkoutResponse, error) {

	profilesReq := user_proto.GetUserRequest{
		Userid: strings.Split(reqUserID.String(), ""), //strings.Split(reqUserID.String(), ""),
	}

	profileRes, err := s.userClient.GetUserProfiles(ctx, &profilesReq)
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
	var wkout []models.Exercise

	exerciseID, err := s.CollectExerciseID(wkout)
	if err != nil {
		//handle err
		log.Printf("cant collect IDs due to : %s", err)
	}

	err = s.ValidateExerciseID(ctx, exerciseID)
	if err != nil {
		//handle err
		log.Printf("issue validating exercise IDs : %v", err)
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
		log.Printf("create work out database Ops error: %s", err)
	}

	var workoutExercisesToInsert []models.WorkoutExerciseResponse
	for _, ex := range exerc {
		workoutExercisesToInsert = append(workoutExercisesToInsert, models.WorkoutExerciseResponse{
			WorkoutID: wkt.ID, // <<< FIX: Use the ID from the record we just created.
			//ExerciseID:  ex.ID,
			ExcerciseID: ex.ID,
			Order:       int32(ex.Order),
			Sets:        int32(ex.Sets),
			Reps:        int32(ex.Reps),
		})
	}

	if err := s.CreateExecrciseRepo(ctx, txs, workoutExercisesToInsert); err != nil {
		log.Printf("iss")
	}
	//if err !=
	if err := txs.Commit(); err != nil {
		log.Printf("Error committing the trasaction to db; %s", err)
	}

	return &models.CreateWorkoutResponse{
		WorkoutID:   wkt.ID,
		Name:        wkt.Name,
		Category:    wkt.Category,
		Description: wkt.Description,
		Exercises:   exerc,
	}, nil
}

func (s *WorkoutService) CreateWorkoutRepo(ctx context.Context, tx *sql.Tx, name string, category string, description string, createdby uuid.UUID) (*models.Workout, error) {

	var createdWorkout models.Workout

	query := `INSERT INTO workouts(name,description,category,created_by)  VALUES($1,$2,$3,$4) RETURNING id,name,description,category,created_by,created_at`

	err := s.db.QueryRowContext(ctx, query, name, description, category, createdby).Scan(
		&createdWorkout.ID,
		&createdWorkout.Name,
		&createdWorkout.Description,
		&createdWorkout.Category,
		&createdWorkout.CreatedBy,
		&createdWorkout.CreatedOn,
		&createdWorkout.UpdatedON,
	)
	if err != nil {
		log.Printf("issue with executing inserting into workout due to: %s", err)
	}

	return &createdWorkout, nil
}

func (s *WorkoutService) CreateExecrciseRepo(ctx context.Context, tx *sql.Tx, exercises []models.WorkoutExerciseResponse) error {

	//bulk insert
	/*
		if len(exercises) == 0 {
			return nil
		}

		sqlStr := `INSERT INTO workout_exercise(workout_id,exercise_id,sequence,sets,reps) VALUES` // ($1,$2,$3,$4,$5)`

		vals := []interface{}{}

		for i, exercise := range exercises {
			placeHolder1 := i*5 + 1
			placeHolder2 := i*5 + 2
			placeHolder3 := i*5 + 3
			placeHolder4 := i*5 + 4
			placeHolder5 := i*5 + 5

			placeholder := fmt.Sprintf("($%d,$%d,$%d,$%d,$%d)", placeHolder1, placeHolder2, placeHolder3, placeHolder4, placeHolder5)

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
		}*/

	if len(exercises) == 0 {
		return nil
	}

	var queryBuilder strings.Builder

	queryBuilder.WriteString(`INSERT INTO workout_exercises (workout_id, exercise_id,sequence, sets, reps) VALUES `)

	const columnCount = 5
	vals := make([]interface{}, 0, len(exercises)*columnCount)

	// --- FIX: The Loop Logic ---
	for i, exercise := range exercises {
		// Calculate the starting parameter index for the current row.
		n := i * columnCount

		// Append the placeholder group for the current row.
		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5))

		// Append a comma if it's NOT the last item in the slice.
		if i < len(exercises)-1 {
			queryBuilder.WriteString(",")
		}

		// Append the actual values to the flat slice.
		vals = append(vals, exercise.WorkoutID, exercise.ExcerciseID, exercise.Order, exercise.Sets, exercise.Reps)
	}

	// Get the final query string from the builder.
	finalQuery := queryBuilder.String()

	// --- EXECUTION ---
	// Execute the query. The 'vals...' unfurls the slice.
	// No need to Prepare, ExecContext can handle it directly.
	result, err := tx.ExecContext(ctx, finalQuery, vals...)
	if err != nil {
		// Log the final query and args for easier debugging if it fails.
		log.Printf("failed to execute bulk insert", "query", finalQuery, "args_count", len(vals), "error", err)
		return fmt.Errorf("failed to execute bulk insert: %w", err)
	}

	// --- VERIFICATION ---
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("could not get rows affected after bulk insert", "error", err)
		return nil // The insert probably succeeded, so we don't return an error.
	}

	if int(rowsAffected) != len(exercises) {
		return fmt.Errorf("bulk insert mismatch: expected %d, inserted %d", len(exercises), int(rowsAffected))
	}

	return nil
}

func (ws *WorkoutService) CollectExerciseID([]models.Exercise) ([]string, error) {
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

type CursorData struct {
	Createdat   time.Time
	WorkoutUUID uuid.UUID
}

var (
	ErrBadRequest = errors.New("invalid data format")
)

func (ws *WorkoutService) ListAllWorkouts(ctx context.Context, reqUserID uuid.UUID, paginationParams models.ListWorkoutParams) (*models.PaginatedWorkout, error) {

	var decodedCursor *CursorData

	if paginationParams.Cursor != " " {
		cursorJSON, err := base64.StdEncoding.DecodeString(paginationParams.Cursor)
		if err != nil {
			log.Printf("issue decoding the cursor parameters to string, %s", err)
			return nil, ErrBadRequest
		}

		var cursorData CursorData

		if err = json.Unmarshal(cursorJSON, &cursorData); err != nil {
			log.Printf("failed to unmarshal the parameter JSON")
			return nil, ErrBadRequest
		}

		decodedCursor = &cursorData
	}

	workouts, err := ws.ListWorkouts(ctx, paginationParams.Limit, paginationParams.Search, decodedCursor)
	if err != nil {
		log.Printf("issue with list workout repository operation due to: %s", err)
		return nil, err
	}

	if len(workouts) == 0 {
		return &models.PaginatedWorkout{
			Data:       []models.Workout{},
			NextCursor: "",
		}, nil
	}

	var nextCursor string

	if len(workouts) == paginationParams.Limit {
		//get last item from the result
		lastWorkout := workouts[len(workouts)-1]

		//create cursor data from the last item
		cursorData := CursorData{
			Createdat:   lastWorkout.CreatedOn,
			WorkoutUUID: lastWorkout.ID,
		}

		//marshal the data into json
		CursorJSON, err := json.Marshal(cursorData)
		if err != nil {
			log.Printf("issue changing the cursor data to JSON, %v", err)
			return nil, err
		}

		nextCursor = base64.StdEncoding.EncodeToString(CursorJSON)
	}

	//asssemble final data

	workoutData := make([]models.Workout, len(workouts))

	for i, workout := range workouts {
		workoutData[i] = models.Workout{
			ID:          workout.ID,
			Name:        workout.Name,
			Description: workout.Description,
			Category:    workout.Category,
			CreatedBy:   workout.CreatedBy,
			CreatedOn:   workout.CreatedOn,
			UpdatedON:   workout.UpdatedON,
		}
	}

	//paginated data
	paginatedResponse := &models.PaginatedWorkout{
		Data:       workoutData,
		NextCursor: nextCursor,
	}

	return paginatedResponse, nil
}

func (ws *WorkoutService) ListWorkouts(ctx context.Context, limit int, search string, cursor *CursorData /*paginationParams models.ListWorkoutParams*/) ([]models.Workout, error) {
	var queryBuilder strings.Builder
	var args []interface{}

	paramIndex := 1

	//base query
	queryBuilder.WriteString("SELECT id,name,description,category,createdby,createdon,updatedon FROM WORKOUT")

	//dynamic where clauses

	whereClause := []string{}

	//search filter
	if search != "" {
		whereClause = append(whereClause, fmt.Sprintf("name ILIKE %d", paramIndex))
		args = append(args, "%"+search+"%")
		paramIndex++
	}

	//cursor pagination
	if cursor != nil {
		clause := fmt.Sprintf("(created_at, uuid) < ($%d, $%d)", paramIndex, paramIndex+1)
		whereClause = append(whereClause, clause)
		args = append(args, cursor.Createdat, cursor.WorkoutUUID)
		paramIndex += 2
	}

	// Append all WHERE clauses if any exist
	if len(whereClause) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(whereClause, " AND "))
	}

	// --- Step 3: Final Clauses ---
	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY created_at DESC, uuid DESC LIMIT $%d", paramIndex))
	args = append(args, limit)

	// --- Step 4 & 5: Execute ---
	finalQuery := queryBuilder.String()
	log.Println("Executing Query:", finalQuery)
	log.Println("With Args:", args)

	rows, err := ws.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	workouts := make([]models.Workout, 0, limit)

	for rows.Next() {
		var workout models.Workout

		if err = rows.Scan(
			&workout.ID,
			&workout.Name,
			&workout.Description,
			&workout.Category,
			&workout.CreatedBy,
			&workout.CreatedOn,
			&workout.UpdatedON,
		); err != nil {
			log.Printf("Issue scanning records, %s", err)
		}

		workouts = append(workouts, workout)
	}

	//workouts  = append(workouts,workout)

	return workouts, nil
}
