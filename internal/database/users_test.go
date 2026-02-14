package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LigeronAhill/luxcarpets-go/internal/database/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersStorage_Create_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	passwordHash := "hashed_password"
	role := types.RoleCustomer
	now := time.Now()

	// ВАЖНО: Добавляем verification_token в список колонок
	mock.ExpectQuery(`INSERT INTO users \(
		    email,
		    username,
		    password_hash,
		    role,
		    image_url,
		    verification_token
		\)
		VALUES \(@email, @username, @password_hash, @role, @image_url, @verification_token\)
		RETURNING \*`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token", // Добавлено поле
		}).AddRow(
			userID, email, false, username, role, nil,
			&passwordHash, now, now, nil, nil, // Добавлено nil для verification_token
		))

	ctx := context.Background()
	params := types.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: &passwordHash,
		Role:         role,
	}

	user, err := storage.Create(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, role, user.Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersStorage_Create_DuplicateEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	email := "duplicate@example.com"
	username := "testuser"
	passwordHash := "hash"
	role := types.RoleCustomer

	// Создаем PgError
	pgErr := &pgconn.PgError{
		Code:           "23505",
		ConstraintName: "users_email_key",
		Message:        "duplicate key value violates unique constraint",
	}

	mock.ExpectQuery(`INSERT INTO users \(
		    email,
		    username,
		    password_hash,
		    role,
		    image_url,
		    verification_token
		\)
		VALUES \(@email, @username, @password_hash, @role, @image_url, @verification_token\)
		RETURNING \*`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(pgErr)

	ctx := context.Background()
	params := types.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: &passwordHash,
		Role:         role,
	}

	user, err := storage.Create(ctx, params)

	// Подробная отладка
	t.Logf("=== DEBUG INFO ===")
	t.Logf("Error type: %T", err)
	t.Logf("Error string: %q", err.Error())

	// Проверяем, можем ли мы получить PgError из цепочки
	var pgErrCheck *pgconn.PgError
	if errors.As(err, &pgErrCheck) {
		t.Logf("SUCCESS: Found PgError in chain")
		t.Logf("  Code: %s", pgErrCheck.Code)
		t.Logf("  Constraint: %s", pgErrCheck.ConstraintName)
	} else {
		t.Logf("FAIL: Could not extract PgError from err")

		// Пробуем разобрать цепочку ошибок вручную
		var unwrapped error = err
		for i := 0; i < 10 && unwrapped != nil; i++ {
			t.Logf("Level %d: %T", i, unwrapped)
			if pgErr, ok := unwrapped.(*pgconn.PgError); ok {
				t.Logf("%v FOUND at level %d!", pgErr, i)
				break
			}
			unwrapped = errors.Unwrap(unwrapped)
		}
	}

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "email already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersStorage_GetByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	role := types.RoleCustomer
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM users WHERE id = @id AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token", // Добавлено поле
		}).AddRow(
			userID, email, true, username, role, nil,
			nil, now, now, nil, nil, // Добавлено nil для verification_token
		))

	ctx := context.Background()
	user, err := storage.GetByID(ctx, userID)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, username, user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersStorage_GetByID_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	userID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM users WHERE id = @id AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	ctx := context.Background()
	user, err := storage.GetByID(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersStorage_GetByEmail_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	role := types.RoleCustomer
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM users WHERE email = @email AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			userID, email, true, username, role, nil,
			nil, now, now, nil, nil,
		))

	ctx := context.Background()
	user, err := storage.GetByEmail(ctx, email)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, username, user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersStorage_Update_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	userID := uuid.New()
	newUsername := "updated_user"
	now := time.Now()

	mock.ExpectQuery(`UPDATE users
		SET
		    username = COALESCE\(@username, username\),
		    role = COALESCE\(@role, role\),
		    image_url = COALESCE\(@image_url, image_url\),
		    email_verified = COALESCE\(@email_verified, email_verified\),
		    verification_token = COALESCE\(@verification_token, verification_token\),
		    password_hash = COALESCE\(@password_hash, password_hash\)
		WHERE id = @id AND deleted_at IS NULL
		RETURNING \*;`).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token", // Добавлено поле
		}).AddRow(
			userID, "test@example.com", false, newUsername, types.RoleCustomer, nil,
			nil, now, now, nil, nil, // Добавлено nil для verification_token
		))

	ctx := context.Background()
	params := types.UpdateUserParams{
		ID:       userID,
		Username: &newUsername,
	}

	user, err := storage.Update(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, newUsername, user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersStorage_List_WithPagination(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	// COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE deleted_at IS NULL`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(5))

	// SELECT query with pagination - ВАЖНО: добавляем все поля
	mock.ExpectQuery(`SELECT \* FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT @limit`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token", // Все поля
		}).AddRow(
			uuid.New(), "user1@test.com", false, "user1", types.RoleCustomer, nil,
			nil, time.Now(), time.Now(), nil, nil,
		).AddRow(
			uuid.New(), "user2@test.com", false, "user2", types.RoleEmployee, nil,
			nil, time.Now(), time.Now(), nil, nil,
		))

	ctx := context.Background()
	params := types.ListUsersParams{
		Limit:  2,
		Offset: 0,
	}

	response, err := storage.List(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 5, response.Total)
	assert.Len(t, response.Data, 2)
	assert.True(t, response.HasNextPage)
	assert.False(t, response.HasPreviousPage)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersStorage_Delete_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	userID := uuid.New()

	mock.ExpectExec(`UPDATE users SET deleted_at = NOW\(\) WHERE id = @id AND deleted_at IS NULL;`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	ctx := context.Background()
	err = storage.Delete(ctx, userID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersStorage_Delete_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	userID := uuid.New()

	mock.ExpectExec(`UPDATE users SET deleted_at = NOW\(\) WHERE id = @id AND deleted_at IS NULL;`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	ctx := context.Background()
	err = storage.Delete(ctx, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Добавим тест для метода GetByEmail, который отсутствовал
func TestUsersStorage_GetByEmail_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := NewUsersStorage(mock)

	email := "nonexistent@example.com"

	mock.ExpectQuery(`SELECT \* FROM users WHERE email = @email AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	ctx := context.Background()
	user, err := storage.GetByEmail(ctx, email)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.NoError(t, mock.ExpectationsWereMet())
}
