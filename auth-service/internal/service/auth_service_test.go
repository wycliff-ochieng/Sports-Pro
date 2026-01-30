package service

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var (
	selectExistsQuery = regexp.QuoteMeta("SELECT EXISTS(SELECT 1 FROM Users WHERE email = $1)")
	insertUserQuery   = regexp.QuoteMeta("INSERT INTO Users(firstname,lastname,email,password,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,userid")
	insertRoleQuery   = regexp.QuoteMeta("INSERT INTO user_roles(user_id,role_id) VALUES($1,$2)")
	selectUserQuery   = regexp.QuoteMeta("SELECT id,userid, email,password,firstname,lastname,created_at,updated_at FROM Users WHERE email = $1")
	selectRolesQuery  = regexp.QuoteMeta("SELECT r.name FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id=$1")
)

func newAuthServiceWithMock(t *testing.T) (*AuthService, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	svc := NewAuthService(db)

	cleanup := func() {
		require.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	}

	return svc, mock, cleanup
}

func TestAuthServiceRegisterSuccess(t *testing.T) {
	svc, mock, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	ctx := context.Background()
	email := "jane.doe@example.com"
	userUUID := uuid.New()

	mock.ExpectQuery(selectExistsQuery).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(insertUserQuery).
		WithArgs("Jane", "Doe", email, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "userid"}).AddRow(101, userUUID))

	mock.ExpectExec(insertRoleQuery).
		WithArgs(101, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	user, err := svc.Register(ctx, "Jane", "Doe", email, "Sup3rSecret!")
	require.NoError(t, err)
	require.Equal(t, userUUID, user.UserID)
	require.Equal(t, "Jane", user.FirstName)
}

func TestAuthServiceRegisterDuplicateEmail(t *testing.T) {
	svc, mock, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	email := "taken@example.com"

	mock.ExpectQuery(selectExistsQuery).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	_, err := svc.Register(context.Background(), "Jane", "Doe", email, "pass")
	require.ErrorIs(t, err, ErrEmailExists)
}

func TestAuthServiceLoginSuccess(t *testing.T) {
	svc, mock, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	ctx := context.Background()
	email := "john.doe@example.com"
	hashed := mustHashPassword(t, "Sup3rSecret!")
	now := time.Now().UTC()
	userUUID := uuid.New()

	mock.ExpectQuery(selectUserQuery).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "userid", "email", "password", "firstname", "lastname", "created_at", "updated_at"}).
			AddRow(88, userUUID, email, hashed, "John", "Doe", now, now))

	mock.ExpectQuery(selectRolesQuery).
		WithArgs(88).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("admin"))

	token, resp, err := svc.Login(ctx, email, "Sup3rSecret!")
	require.NoError(t, err)
	require.NotNil(t, token)
	require.Equal(t, "John", resp.FirstName)
}

func TestAuthServiceLoginInvalidPassword(t *testing.T) {
	svc, mock, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	email := "johnny@example.com"
	hashed := mustHashPassword(t, "correctpass")
	now := time.Now().UTC()

	mock.ExpectQuery(selectUserQuery).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "userid", "email", "password", "firstname", "lastname", "created_at", "updated_at"}).
			AddRow(5, uuid.New(), email, hashed, "John", "Doe", now, now))

	_, _, err := svc.Login(context.Background(), email, "badpass")
	require.Error(t, err)
	require.Contains(t, err.Error(), "passwords do no match")
}

func TestAuthServiceLoginUserNotFound(t *testing.T) {
	svc, mock, cleanup := newAuthServiceWithMock(t)
	defer cleanup()

	email := "missing@example.com"

	mock.ExpectQuery(selectUserQuery).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	_, _, err := svc.Login(context.Background(), email, "whatever")
	require.ErrorIs(t, err, ErrNotFound)
}

func mustHashPassword(t *testing.T, password string) string {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(hashed)
}
