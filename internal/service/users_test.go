package service

import (
	"context"
	"testing"
	"time"

	"github.com/LigeronAhill/luxcarpets-go/internal/database"
	"github.com/LigeronAhill/luxcarpets-go/internal/database/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsersService_SignUp_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	password := "TestP@ssw0rd"
	role := string(types.RoleGuest)
	imageURL := "https://example.com/avatar.jpg"
	verificationToken := "token123"
	now := time.Now()

	// Используем AnyArg() для всех параметров
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
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			userID, email, false, username, types.RoleGuest, &imageURL,
			nil, now, now, nil, &verificationToken,
		))

	ctx := context.Background()
	result, err := service.SignUp(ctx, email, username, &password, &role, &imageURL, &verificationToken)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, username, result.Username)
	assert.Equal(t, types.RoleGuest, result.Role)
	assert.Equal(t, imageURL, result.ImageURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_SignUp_InvalidPassword(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	email := "test@example.com"
	username := "testuser"
	password := "weak"

	ctx := context.Background()
	result, err := service.SignUp(ctx, email, username, &password, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid password")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_SignUp_InvalidRole(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	email := "test@example.com"
	username := "testuser"
	password := "TestP@ssw0rd"
	invalidRole := "invalid_role"

	ctx := context.Background()
	result, err := service.SignUp(ctx, email, username, &password, &invalidRole, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid role")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_SignIn_Password_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	password := "TestP@ssw0rd"
	hashedPassword, _ := hashPassword(password)
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM users WHERE email = @email AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			userID, email, false, username, types.RoleGuest, nil,
			&hashedPassword, now, now, nil, nil,
		))

	ctx := context.Background()
	result, err := service.SignIn(ctx, email, &password, nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, username, result.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_SignIn_Token_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	token := "valid-token"
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM users WHERE email = @email AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			userID, email, false, username, types.RoleGuest, nil,
			nil, now, now, nil, &token,
		))

	ctx := context.Background()
	result, err := service.SignIn(ctx, email, nil, &token)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, username, result.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_SignIn_EmailRequired(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	ctx := context.Background()
	result, err := service.SignIn(ctx, "", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrEmailRequired)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_SignIn_NoCredentials(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	// Добавляем мок для запроса, так как метод сначала ищет пользователя
	mock.ExpectQuery(`SELECT \* FROM users WHERE email = @email AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			uuid.New(), "test@example.com", false, "testuser", types.RoleGuest, nil,
			nil, time.Now(), time.Now(), nil, nil,
		))

	ctx := context.Background()
	result, err := service.SignIn(ctx, "test@example.com", nil, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrPasswordOrTokenReq)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_SignIn_UserNotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	email := "nonexistent@example.com"
	password := "TestP@ssw0rd"

	mock.ExpectQuery(`SELECT \* FROM users WHERE email = @email AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(pgx.ErrNoRows)

	ctx := context.Background()
	result, err := service.SignIn(ctx, email, &password, nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_GetByID_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM users WHERE id = @id AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			userID, email, false, username, types.RoleGuest, nil,
			nil, now, now, nil, nil,
		))

	ctx := context.Background()
	result, err := service.GetByID(ctx, userID.String())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, username, result.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_GetByID_InvalidID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	ctx := context.Background()
	result, err := service.GetByID(ctx, "invalid-uuid")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrUserIDRequired)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_GetByEmail_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	userID := uuid.New()
	email := "test@example.com"
	username := "testuser"
	now := time.Now()

	mock.ExpectQuery(`SELECT \* FROM users WHERE email = @email AND deleted_at IS NULL`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			userID, email, false, username, types.RoleGuest, nil,
			nil, now, now, nil, nil,
		))

	ctx := context.Background()
	result, err := service.GetByEmail(ctx, email)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, email, result.Email)
	assert.Equal(t, username, result.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_Update_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

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
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			userID, "test@example.com", false, newUsername, types.RoleGuest, nil,
			nil, now, now, nil, nil,
		))

	ctx := context.Background()
	params := types.UpdateUserParams{
		ID:       userID,
		Username: &newUsername,
	}

	result, err := service.Update(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newUsername, result.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_Update_InvalidPassword(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	userID := uuid.New()
	weakPassword := "weak"

	ctx := context.Background()
	params := types.UpdateUserParams{
		ID:           userID,
		PasswordHash: &weakPassword,
	}

	result, err := service.Update(ctx, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrPasswordTooShort)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_List_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	// COUNT query
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE deleted_at IS NULL`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))

	// SELECT query
	mock.ExpectQuery(`SELECT \* FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT @limit`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "email", "email_verified", "username", "role", "image_url",
			"password_hash", "created_at", "updated_at", "deleted_at", "verification_token",
		}).AddRow(
			uuid.New(), "user1@test.com", false, "user1", types.RoleGuest, nil,
			nil, time.Now(), time.Now(), nil, nil,
		).AddRow(
			uuid.New(), "user2@test.com", false, "user2", types.RoleAdmin, nil,
			nil, time.Now(), time.Now(), nil, nil,
		))

	ctx := context.Background()
	params := types.ListUsersParams{
		Limit:  10,
		Offset: 0,
	}

	result, err := service.List(ctx, params)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, len(result.Data))
	assert.Equal(t, 2, result.Total)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_List_InvalidOffset(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	ctx := context.Background()
	params := types.ListUsersParams{
		Limit:  10,
		Offset: -1,
	}

	result, err := service.List(ctx, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidOffset)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_Delete_Success(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	userID := uuid.New()

	mock.ExpectExec(`UPDATE users SET deleted_at = NOW\(\) WHERE id = @id AND deleted_at IS NULL;`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	ctx := context.Background()
	err = service.Delete(ctx, userID.String())

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUsersService_Delete_NotFound(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	storage := database.NewUsersStorage(mock)
	service := NewUsersService(storage)

	userID := uuid.New()

	mock.ExpectExec(`UPDATE users SET deleted_at = NOW\(\) WHERE id = @id AND deleted_at IS NULL;`).
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	ctx := context.Background()
	err = service.Delete(ctx, userID.String())

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
