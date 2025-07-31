package service

import (
	"context"
	"database/sql"

	//"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/wycliff-ochieng/internal/database"
	"github.com/wycliff-ochieng/internal/models"
	internal "github.com/wycliff-ochieng/internal/producer"
)

type Profile interface {
	GetProfileByID(ctx context.Context, tx *sql.Tx, userID int) (*models.Profile, error)
	UpdateUserProfile(ctx context.Context, userID int, firstname, lastname, email string, updatedat time.Time) (*models.Profile, error)
}

type UserService struct {
	l  *log.Logger
	db database.DBInterface
	p  internal.KafkaProducer
}

type EventsData struct {
	UserID    int    `json:"userid"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
}

type UpdatePofileReq struct {
	UserID    int
	Firstname string
	Lastname  string
	//avatar    string
	Email string
	//Createdat string
	Updatedat time.Time
}

func NewUserService(l *log.Logger, db database.DBInterface, p internal.KafkaProducer) *UserService {
	return &UserService{
		db: db,
		l:  l,
		p:  p,
		//dbTx:dbTx,
	}
}

var (
	ErrNotFound     = errors.New("usser Id not found")
	ErrUpdateFailed = errors.New("failed to update user profile")
)

func (u *UserService) CreateUserProfile(ctx context.Context, userID int, firstname string, lastname string, email string) error {

	//var event EventsData

	query := `INSERT INTO user_profiles(userid,firstname,lastname,email) VALUES($1,$2,$3,$4)`

	_, err := u.db.ExecContext(ctx, query, userID, firstname, lastname, email)
	if err != nil {
		return fmt.Errorf("something happened: %v", err)
	}

	return nil
}

func (u *UserService) GetProfileByID(ctx context.Context, tx *sql.Tx, userID int) (*models.Profile, error) {
	var event EventsData

	query := "SELECT * from user_profiles WHERE id=$1"

	_, err := u.db.ExecContext(ctx, query, event.UserID)
	if err != nil {
		return nil, err
	}

	return &models.Profile{
		UserID: event.UserID,
		Email:  event.Email,
	}, nil
}

func (u *UserService) UpdateUserProfile(ctx context.Context, userID int, firstname string, lastname string, email string, updatedat time.Time) (*models.Profile, error) {

	u.l.Printf("UPDATING USER PROFILE UNDERWAY FOR:%v", userID)

	var updateReq UpdatePofileReq

	// Begin transaction
	transaction, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		u.l.Printf("Transaction undergone some errors:%v", err)
	}

	//defer rollback - if anything below panicks everything will be roollback
	defer transaction.Rollback()

	//check if user exists, call get user by ID
	existingProfile, err := u.GetProfileByID(ctx, transaction, userID)
	if err != nil {
		u.l.Printf("no user: %d with that ID: %v", userID, err)
		return nil, ErrUpdateFailed
	}

	if existingProfile == nil {
		u.l.Printf("no information about that user")
		return nil, ErrNotFound
	}

	updateReq.UserID = userID

	err = u.UpdateUserProfileRepo(ctx, transaction, updateReq)
	if err != nil {
		u.l.Printf("failed to update profile for user:%s: of ID:%v", updateReq.Firstname, updateReq.UserID)
		return nil, ErrUpdateFailed
	}

	u.l.Printf("successfully update profile for user: %s,%s", firstname, lastname)

	//event := u.UpdateUserProfile()

	//Side effects publish event(Implementing kafka producer)
	err = u.p.PublishUserUpdate(ctx, updateReq.UserID)
	if err != nil {
		u.l.Printf("error publishing update event to producer:%v", err)
		return nil, err
	}

	return &models.Profile{}, nil

}

func (u *UserService) UpdateUserProfileRepo(ctx context.Context, tx *sql.Tx, updatedEventsData UpdatePofileReq) error {

	u.l.Println(">>>>Write operation has started successfully")

	var profile models.Profile

	query := `UPDATE user_profile
	SET firstname=$1,lastname=$2,updatedat=NOW() WHERE userID=$4`

	result, err := tx.ExecContext(ctx, query, profile.Firstname, profile.Lastname, profile.Updatedat)
	if err != nil {
		u.l.Println("Error executing transactions")
		return err
	}

	//check rows affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		u.l.Println("REPO ERROR, failed to check rows affected for user")
		return err
	}

	if rowsAffected == 0 {
		u.l.Println("REPO ERROR , There is not user with that id")
	}

	u.l.Printf("successfully updated profile for : %v", profile.UserID)
	return nil
}

func (u *UserService) GetProfileByIDRepo(ctx context.Context, userID int) (*models.Profile, error) {
	//Redis and caching
	u.l.Println(">>get user by id started successfully")

	var profile models.Profile

	query := `SELECT userId,firstname,lastname,email,createdat,updatedat FROM user_profiles WHERE userId = $1`

	err := u.db.QueryRowContext(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.Firstname,
		&profile.Lastname,
		&profile.Email,
		&profile.Createdat,
		&profile.Updatedat,
	)
	if err == sql.ErrNoRows {
		u.l.Println("no user with that IDin the database")
	}
	return &profile, nil
}
