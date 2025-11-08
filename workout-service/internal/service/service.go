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
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/wycliff-ochieng/internal/database"

	//"github.com/wycliff-ochieng/internal/handlers"
	"github.com/wycliff-ochieng/internal/models"
	"github.com/wycliff-ochieng/sports-common-package/user_grpc/user_proto"
)

var (
	ErrForbidden = errors.New("Not allowed to do this")
	//ErrBadRequest = errors.New("bad r")
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

func (s *WorkoutService) CreateWorkout(ctx context.Context, reqUserID uuid.UUID, req models.CreateWorkoutResponse) (*models.CreateWorkoutResponse, error) {

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

	//begin transaction
	txs, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transactions: %s", err)
		return nil, err
	}

	defer txs.Rollback()

	//validate exercise UUID
	uniqueUUIDs := make(map[uuid.UUID]struct{})
	for _, ex := range req.Exercises {
		uniqueUUIDs[ex.ExerciseID] = struct{}{}
	}

	// Create a slice to pass to the repository function.
	uuidsToValidate := make([]uuid.UUID, 0, len(uniqueUUIDs))
	for u := range uniqueUUIDs {
		uuidsToValidate = append(uuidsToValidate, u)
	}

	// Handle the edge case where no exercises were provided.
	if len(uuidsToValidate) == 0 {
		return nil, ErrBadRequest
	}

	//err = s.ValidateExerciseUUIDs(ctx, txs, uuidsToValidate)
	//if err != nil {
	//	log.Println("exercise UUID validation failed", "error", err)

	// Wrap the repository error in a more specific service-level error.
	//	return nil, ErrBadRequest
	//}

	workoutToCreate := &models.Workout{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		CreatedBy:   reqUserID,
	}

	newlyCreatedWorkout, err := s.CreateWorkoutRepo(ctx, txs, workoutToCreate)
	if err != nil {
		log.Println("CreateWorkoutRepo failed", "error", err)
		return nil, err
	}

	linksToCreate := make([]WorkoutExerciseLink, 0, len(req.Exercises))

	// Loop through the exercises from the original client request.
	for _, exerciseFromReq := range req.Exercises {
		// For each exercise, create a link object.
		link := WorkoutExerciseLink{
			WorkoutID:  newlyCreatedWorkout.ID, // <-- CRITICAL: Use the ID from the workout we just created.
			ExerciseID: exerciseFromReq.ExerciseID,
			Order:      int(exerciseFromReq.Order),
			Sets:       int(exerciseFromReq.Sets),
			Reps:       int(exerciseFromReq.Reps), //.String(),
		}

		linksToCreate = append(linksToCreate, link)
	}

	// --- Step 3: Call the Second Repository Function ---
	// Concept: Perform the bulk insert of all the link records.
	if len(linksToCreate) > 0 {
		err := s.CreateBulkWorkoutExercises(ctx, txs, linksToCreate)
		if err != nil {
			// If this fails, stop and return the error.
			log.Println("CreateBulkWorkoutExercises failed", "error", err)
			return nil, err
		}
	}

	if err := txs.Commit(); err != nil {
		log.Printf("Error committing the trasaction to db; %s", err)
		return nil, err
	}

	return &models.CreateWorkoutResponse{
		WorkoutID:   newlyCreatedWorkout.ID,
		Name:        newlyCreatedWorkout.Name,
		Category:    newlyCreatedWorkout.Category,
		Description: newlyCreatedWorkout.Description,
		Exercises:   req.Exercises,
	}, nil

}

func (s *WorkoutService) CreateWorkoutRepo(ctx context.Context, tx *sql.Tx, workout *models.Workout) (*models.Workout, error) {
	// This struct will hold the data scanned back from the database.
	var createdWorkout models.Workout

	query := `
        INSERT INTO workouts (name, description, category, created_by)
        VALUES ($1, $2, $3, $4)
        RETURNING id, name, description, category, created_by, created_at, updated_at
    `

	err := tx.QueryRowContext(ctx, query,
		workout.Name,
		workout.Description,
		workout.Category,
		workout.CreatedBy,
	).Scan(
		&createdWorkout.ID,
		&createdWorkout.Name,
		&createdWorkout.Description,
		&createdWorkout.Category,
		&createdWorkout.CreatedBy,
		&createdWorkout.CreatedOn, // Ensure these field names match your Go struct
		&createdWorkout.UpdatedON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert workout record: %w", err)
	}

	return &createdWorkout, nil
}

type WorkoutExerciseLink struct {
	WorkoutID  uuid.UUID
	ExerciseID uuid.UUID
	Order      int
	Sets       int
	Reps       int //string // Use string for flexibility
}

func (r *WorkoutService) CreateBulkWorkoutExercises(ctx context.Context, tx *sql.Tx, links []WorkoutExerciseLink) error {
	if len(links) == 0 {
		return nil // Nothing to do.
	}

	// This function uses the standard database/sql approach for bulk inserts.
	var queryBuilder strings.Builder
	queryBuilder.WriteString(`INSERT INTO workout_exercises (workout_id, exercise_id, "sequence", sets, reps) VALUES `)

	const columnCount = 5
	vals := make([]interface{}, 0, len(links)*columnCount)

	for i, link := range links {
		n := i * columnCount
		queryBuilder.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", n+1, n+2, n+3, n+4, n+5))

		if i < len(links)-1 {
			queryBuilder.WriteString(",")
		}

		vals = append(vals, link.WorkoutID, link.ExerciseID, link.Order, link.Sets, link.Reps)
	}

	finalQuery := queryBuilder.String()

	// Execute the single, large INSERT statement on the transaction.
	_, err := tx.ExecContext(ctx, finalQuery, vals...)
	if err != nil {
		return fmt.Errorf("failed to execute bulk insert for workout_exercises: %w", err)
	}

	return nil
}

func (s *WorkoutService) ValidateExerciseUUIDs(ctx context.Context, tx *sql.Tx, exerciseUUIDs []uuid.UUID) error {
	if len(exerciseUUIDs) == 0 {
		return nil // Nothing to validate.
	}

	// This is the validation query. It's highly efficient for PostgreSQL.
	query := `SELECT COUNT(id) FROM exercises WHERE id = ANY($1)`

	// The pq.Array helper is used to properly format the Go slice of UUIDs
	// into a format that PostgreSQL's array parameter type understands.
	var count int
	err := tx.QueryRowContext(ctx, query, pq.Array(exerciseUUIDs)).Scan(&count)
	if err != nil {
		// This indicates a problem with the query itself or the database connection.
		return fmt.Errorf("failed to execute exercise validation query: %w", err)
	}

	// Compare the count of found UUIDs with the number of unique UUIDs we checked for.
	if count != len(exerciseUUIDs) {
		// This is the validation failure. At least one UUID was not found.
		// We return a specific, recognizable error.
		return fmt.Errorf("validation failed: one or more exercise UUIDs do not exist (found %d, expected %d)", count, len(exerciseUUIDs))
	}

	// If the counts match, all UUIDs are valid.
	return nil
}

/*
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
		return nil, fmt.Errorf("error: %s", err)
	}

	err = s.ValidateExerciseID(ctx, exerciseID)
	if err != nil {
		//handle err
		log.Printf("issue validating exercise IDs : %v", err)
		return nil, err
	}

	/*seenUUIDs := make(map[uuid.UUID]bool)
	for _, ex := range exerc {
		if seenUUIDs[ex.ID] {
			// We have found a duplicate in the request itself.
			return nil, ErrBadRequest
		}
		seenUUIDs[ex.ID] = true
	}*

	//begin transaction
	txs, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transactions: %s", err)
		return nil, err
	}

	defer txs.Rollback()

	//newWorkoutID := wkout.ID

	//exerciseQuery := `INSERT INTO workout_exercise() VALUES($1,$2,$3,$4)`

	//err = txs.QueryRowContext(ctx, exerciseQuery).Scan()

	wkt, err := s.CreateWorkoutRepo(ctx, txs, name, category, description, reqUserID)
	if err != nil {
		//handle error
		log.Printf("create work out database Ops error: %s", err)
		return nil, err
	}

	var workoutExercisesToInsert []models.WorkoutExerciseResponse
	for _, ex := range exerc {
		workoutExercisesToInsert = append(workoutExercisesToInsert, models.WorkoutExerciseResponse{
			WorkoutID:   wkt.ID, // <<< FIX: Use the ID from the record we just created.
			ExcerciseID: ex.ID,
			Order:       int32(ex.Order),
			Sets:        int32(ex.Sets),
			Reps:        int32(ex.Reps),
		})
	}

	if err := s.CreateExecrciseRepo(ctx, txs, workoutExercisesToInsert); err != nil {
		log.Printf("an erro due to: %s", err)
		return nil, err
	}
	//if err !=
	if err := txs.Commit(); err != nil {
		log.Printf("Error committing the trasaction to db; %s", err)
		return nil, err
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

	err := tx.QueryRowContext(ctx, query, name, description, category, createdby).Scan(
		&createdWorkout.ID,
		&createdWorkout.Name,
		&createdWorkout.Description,
		&createdWorkout.Category,
		&createdWorkout.CreatedBy,
		&createdWorkout.CreatedOn,
		//&createdWorkout.UpdatedON,
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
		}*

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
		log.Print("failed to execute bulk insert", "query", finalQuery, "args_count", len(vals), "error", err)
		return fmt.Errorf("failed to execute bulk insert: %w", err)
	}

	// --- VERIFICATION ---
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Print("could not get rows affected after bulk insert", "error", err)
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

*/

type CursorData struct {
	Createdat   time.Time
	WorkoutUUID uuid.UUID
}

var (
	ErrBadRequest = errors.New("invalid data format")
)

func (ws *WorkoutService) ListAllWorkouts(ctx context.Context, reqUserID uuid.UUID, paginationParams models.ListWorkoutParams) (*models.PaginatedWorkout, error) {

	var decodedCursor *CursorData

	if paginationParams.Cursor != "" {
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
	queryBuilder.WriteString("SELECT id,name,description,category,created_by,created_at,updated_at FROM workouts")

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
		clause := fmt.Sprintf("(created_at, id) < ($%d, $%d)", paramIndex, paramIndex+1)
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
	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d", paramIndex))
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

func (ws *WorkoutService) CreateExercise(ctx context.Context, reqUser uuid.UUID, req *models.CreateExerciseReq) (*models.Exercise, error) {
	var ex models.Exercise

	exercise, err := models.NewExercise(req.Name, req.Description, req.Instructions)
	if err != nil {
		return nil, fmt.Errorf("issue creating exercise: %s", err)
	}

	query := `INSERT INTO exercise(id,name,description,instructions,created_by,created_at,updated_at)VALUES($1,$2,$3,$4,$5,$6,$7)`

	_, err = ws.db.ExecContext(ctx, query, ex.ID, ex.Name, ex.Description, ex.Instruction, ex.CreatedBy, ex.CreatedOn, ex.UpdatedOn)
	if err != nil {
		return nil, fmt.Errorf("issue creating exercise due to: %s", err)
	}

	return &models.Exercise{
		ID:          uuid.New(),
		Name:        exercise.Name,
		Description: exercise.Description,
		Instruction: exercise.Instruction,
		CreatedBy:   reqUser,
	}, nil
}

//func(ws *WorkoutService) CreateWorkoutRepo(ctx, )
