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
	"github.com/wycliff-ochieng/internal/filestore"

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
	mnClient   *filestore.FileStore
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

	query := `SELECT COUNT(id) FROM exercises WHERE id = ANY($1)`

	var count int
	err := tx.QueryRowContext(ctx, query, pq.Array(exerciseUUIDs)).Scan(&count)
	if err != nil {
		// This indicates a problem with the query itself or the database connection.
		return fmt.Errorf("failed to execute exercise validation query: %w", err)
	}

	if count != len(exerciseUUIDs) {

		return fmt.Errorf("validation failed: one or more exercise UUIDs do not exist (found %d, expected %d)", count, len(exerciseUUIDs))
	}

	// If the counts match, all UUIDs are valid.
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

func (ws *WorkoutService) CreateExercise(ctx context.Context, reqUser uuid.UUID, name, description, instructions string) (*models.Exercise, error) {
	var ex models.Exercise

	/*exercise, err := models.NewExercise(name, description, instructions)
	if err != nil {
		log.Printf("error creating exercise due to: %s", err)
		return nil, fmt.Errorf("issue creating exercise: %s", err)
	}*/

	query := `INSERT INTO exercises(name,description,instructions,created_by)VALUES($1,$2,$3,$4) RETURNING 
	id,name,description,instructions,created_by,created_at,updated_at`

	err := ws.db.QueryRowContext(ctx, query, name, description, instructions, reqUser).Scan(
		&ex.ID,
		&ex.Name,
		&ex.Description,
		&ex.Instruction,
		&ex.CreatedBy,
		&ex.CreatedOn,
		&ex.UpdatedOn,
	)
	if err != nil {
		log.Printf("error executing due to: %s", err)
		return nil, fmt.Errorf("issue creating exercise due to: %s", err)
	}

	return &models.Exercise{
		ID:          ex.ID,
		Name:        ex.Name,
		Description: ex.Description,
		Instruction: ex.Instruction,
		CreatedBy:   reqUser,
	}, nil
}

func (ws *WorkoutService) GetExercises(ctx context.Context, reqUser uuid.UUID) (*[]models.Exercise, error) {
	var allExercises []models.Exercise

	query := `SELECT id,name,description,instructions,created_by,created_at,updated_at FROM exercises;`

	rows, err := ws.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("issue fetching exercises due to: %s", err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {

		var exercises models.Exercise

		err := rows.Scan(
			&exercises.ID,
			&exercises.Name,
			&exercises.Description,
			&exercises.Instruction,
			&exercises.CreatedBy,
			&exercises.CreatedOn,
			&exercises.UpdatedOn,
		)
		if err != nil {
			log.Printf("issue fething exercises in row next due to: %s", err)
		}

		allExercises = append(allExercises, exercises)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &allExercises, nil
}

// frontend->Go(Backend)->MiniIO
func (ws *WorkoutService) GeneratePresignedURL(ctx context.Context, req *models.PresignedURLReq) (*models.PresignedURLRes, error) {

	//create patth
	filePath := fmt.Sprintf("%s/%s/%s-%s", req.ParentType, req.ParentID, uuid.New(), req.Filename)

	//connect /talk to MinIO
	url, err := ws.mnClient.Client.PresignedPutObject(ctx, bucket, filePath, 10*time.Second)
	if err != nil {
		fmt.Errorf("error generating url due to :%s", err)
	}
}
