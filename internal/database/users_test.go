package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/LigeronAhill/luxcarpets-go/internal/database/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type UsersStorageTestSuite struct {
	suite.Suite
	ctx      context.Context
	pool     *pgxpool.Pool
	storage  *UsersStorage
	cleanup  func()
	testUser *types.User
}

func TestUsersStorageSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration tests in short mode")
	}

	suite.Run(t, new(UsersStorageTestSuite))
}

func (s *UsersStorageTestSuite) SetupSuite() {
	s.ctx = context.Background()

	// Получаем URL тестовой БД из переменной окружения
	dbURL := os.Getenv("LUXCARPETS_DATABASE_TESTURL")
	if dbURL == "" {
		// Проверяем разные варианты подключения
		dbURLs := []string{
			"postgres://postgres:postgres@localhost:5432/luxcarpets_test",
			"postgres://postgres:postgres@localhost:5433/luxcarpets_test",
		}

		for _, url := range dbURLs {
			s.T().Logf("Trying database URL: %s", url)
			pool, err := pgxpool.New(s.ctx, url)
			if err == nil {
				err = pool.Ping(s.ctx)
				if err == nil {
					dbURL = url
					pool.Close()
					break
				}
				pool.Close()
			}
		}

		if dbURL == "" {
			s.T().Fatal("No database connection available. Set LUXCARPETS_DATABASE_TESTURL environment variable.")
		}
	}

	s.T().Logf("Using database URL: %s", dbURL)

	// Создаем пул подключений
	s.pool = NewPool(s.ctx, dbURL)
	s.storage = NewUsersStorage(s.pool)

	// Очистка таблицы перед тестами
	s.cleanup = func() {
		_, err := s.pool.Exec(s.ctx, "DELETE FROM users")
		if err != nil {
			s.T().Logf("Warning: failed to clean up users table: %v", err)
		}
	}
}

func (s *UsersStorageTestSuite) SetupTest() {
	s.cleanup()

	// Создаем тестового пользователя для использования в тестах
	createParams := types.CreateUserParams{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: stringPtr("hashed_password_123"),
		Role:         types.RoleCustomer,
	}

	result := s.storage.Create(s.ctx, createParams)

	// Добавляем отладочную информацию
	if result.IsErr() {
		s.T().Logf("ERROR creating test user: %v", result.Error)

		// Проверяем схему БД
		var tableExists bool
		err := s.pool.QueryRow(s.ctx,
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&tableExists)
		s.T().Logf("Table 'users' exists: %v", tableExists)

		if err != nil {
			s.T().Logf("Error checking table existence: %v", err)
		}

		// Проверяем структуру таблицы
		rows, err := s.pool.Query(s.ctx,
			"SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'users' ORDER BY ordinal_position")
		if err != nil {
			s.T().Logf("Error checking table structure: %v", err)
		} else {
			defer rows.Close()
			s.T().Log("Table 'users' columns:")
			for rows.Next() {
				var colName, dataType string
				rows.Scan(&colName, &dataType)
				s.T().Logf("  - %s: %s", colName, dataType)
			}
		}
	}

	require.True(s.T(), result.IsOk(), "Failed to create test user: %v", result.Error)

	s.testUser = result.Must()
	s.T().Logf("Created test user with ID: %s", s.testUser.ID)
}

func (s *UsersStorageTestSuite) TearDownSuite() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// TestCreateUser тестирует создание пользователя
func (s *UsersStorageTestSuite) TestCreateUser() {
	t := s.T()

	tests := []struct {
		name     string
		params   types.CreateUserParams
		wantErr  bool
		checkErr func(error) bool
	}{
		{
			name: "Успешное создание пользователя",
			params: types.CreateUserParams{
				Email:        "newuser@example.com",
				Username:     "newuser",
				PasswordHash: stringPtr("hashed_password"),
				Role:         types.RoleCustomer,
			},
			wantErr: false,
		},
		{
			name: "Создание пользователя без пароля",
			params: types.CreateUserParams{
				Email:    "nopassword@example.com",
				Username: "nopassword",
				Role:     types.RoleGuest,
			},
			wantErr: false,
		},
		{
			name: "Дубликат email",
			params: types.CreateUserParams{
				Email:        s.testUser.Email, // Используем email существующего пользователя
				Username:     "differentuser",
				PasswordHash: stringPtr("password"),
				Role:         types.RoleCustomer,
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return err.Error() == "email already exists"
			},
		},
		{
			name: "Дубликат username",
			params: types.CreateUserParams{
				Email:        "different@example.com",
				Username:     s.testUser.Username, // Используем username существующего пользователя
				PasswordHash: stringPtr("password"),
				Role:         types.RoleCustomer,
			},
			wantErr: false,
			checkErr: func(err error) bool {
				return IsUniqueConstraintViolation(err, "users_username_key")
			},
		},
		{
			name: "Некорректный email",
			params: types.CreateUserParams{
				Email:        "invalid-email",
				Username:     "invaliduser",
				PasswordHash: stringPtr("password"),
				Role:         types.RoleCustomer,
			},
			wantErr: true,
			checkErr: func(err error) bool {
				var pgErr *pgconn.PgError
				return errors.As(err, &pgErr) && pgErr.ConstraintName == "chk_email"
			},
		},
		{
			name: "Слишком короткий username",
			params: types.CreateUserParams{
				Email:        "short@example.com",
				Username:     "ab", // Менее 3 символов
				PasswordHash: stringPtr("password"),
				Role:         types.RoleCustomer,
			},
			wantErr: true,
			checkErr: func(err error) bool {
				var pgErr *pgconn.PgError
				return errors.As(err, &pgErr) && pgErr.ConstraintName == "chk_username_length"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.storage.Create(s.ctx, tt.params)

			if tt.wantErr {
				assert.True(t, result.IsErr(), "Expected error but got success")
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(result.Error),
						"Error doesn't match expected: %v", result.Error)
				}
			} else {
				assert.True(t, result.IsOk(), "Expected success but got error: %v", result.Error)

				user := result.Must()
				assert.NotEqual(t, uuid.Nil, user.ID)
				assert.Equal(t, tt.params.Email, user.Email)
				assert.Equal(t, tt.params.Username, user.Username)
				assert.Equal(t, tt.params.Role, user.Role)
				assert.False(t, user.EmailVerified)
				assert.NotZero(t, user.CreatedAt)
				assert.NotZero(t, user.UpdatedAt)

				// Проверяем значения по умолчанию
				if tt.params.PasswordHash != nil {
					assert.NotNil(t, user.PasswordHash)
					assert.Equal(t, *tt.params.PasswordHash, *user.PasswordHash)
				} else {
					assert.Nil(t, user.PasswordHash)
				}

				if tt.params.ImageURL != nil {
					assert.NotNil(t, user.ImageURL)
					assert.Equal(t, *tt.params.ImageURL, *user.ImageURL)
				} else {
					assert.Nil(t, user.ImageURL)
				}

				if tt.params.VerificationToken != nil {
					assert.NotNil(t, user.VerificationToken)
					assert.Equal(t, *tt.params.VerificationToken, *user.VerificationToken)
				} else {
					assert.Nil(t, user.VerificationToken)
				}

				// DeletedAt должен быть nil для нового пользователя
				assert.Nil(t, user.DeletedAt)
			}
		})
	}
}

// TestGetByID тестирует получение пользователя по ID
func (s *UsersStorageTestSuite) TestGetByID() {
	t := s.T()

	t.Run("Успешное получение существующего пользователя", func(t *testing.T) {
		result := s.storage.GetByID(s.ctx, s.testUser.ID)
		assert.True(t, result.IsOk(), "Failed to get user: %v", result.Error)

		user := result.Must()
		assert.Equal(t, s.testUser.ID, user.ID)
		assert.Equal(t, s.testUser.Email, user.Email)
		assert.Equal(t, s.testUser.Username, user.Username)
		assert.Equal(t, s.testUser.Role, user.Role)
	})

	t.Run("Получение несуществующего пользователя", func(t *testing.T) {
		nonExistentID := uuid.New()
		result := s.storage.GetByID(s.ctx, nonExistentID)
		assert.True(t, result.IsErr(), "Expected error for non-existent user")
	})

	t.Run("Получение удаленного пользователя", func(t *testing.T) {
		// Создаем и удаляем пользователя
		createParams := types.CreateUserParams{
			Email:    "todelete@example.com",
			Username: "todelete",
			Role:     types.RoleCustomer,
		}

		createResult := s.storage.Create(s.ctx, createParams)
		require.True(t, createResult.IsOk())

		userToDelete := createResult.Must()

		// Удаляем пользователя
		err := s.storage.Delete(s.ctx, userToDelete.ID)
		require.NoError(t, err)

		// Пытаемся получить удаленного пользователя
		result := s.storage.GetByID(s.ctx, userToDelete.ID)
		assert.True(t, result.IsErr(), "Expected error for deleted user")
	})
}

// TestGetByEmail тестирует получение пользователя по email
func (s *UsersStorageTestSuite) TestGetByEmail() {
	t := s.T()

	t.Run("Успешное получение по email", func(t *testing.T) {
		result := s.storage.GetByEmail(s.ctx, s.testUser.Email)
		assert.True(t, result.IsOk())

		user := result.Must()
		assert.Equal(t, s.testUser.ID, user.ID)
		assert.Equal(t, s.testUser.Email, user.Email)
	})

	t.Run("Получение по несуществующему email", func(t *testing.T) {
		result := s.storage.GetByEmail(s.ctx, "nonexistent@example.com")
		assert.True(t, result.IsErr())
	})

	t.Run("Регистронезависимый поиск", func(t *testing.T) {
		// Создаем пользователя с email в разных регистрах
		email := "MixedCase@Example.COM"
		createParams := types.CreateUserParams{
			Email:    email,
			Username: "mixedcase",
			Role:     types.RoleCustomer,
		}

		createResult := s.storage.Create(s.ctx, createParams)
		require.True(t, createResult.IsOk())

		// Ищем в нижнем регистре
		result := s.storage.GetByEmail(s.ctx, "mixedcase@example.com")
		assert.True(t, result.IsOk())

		user := result.Must()
		assert.Equal(t, strings.ToLower(email), user.Email)
	})
}

// TestUpdate тестирует обновление пользователя
func (s *UsersStorageTestSuite) TestUpdate() {
	t := s.T()

	t.Run("Успешное обновление всех полей", func(t *testing.T) {
		newUsername := "updateduser"
		newRole := types.RoleAdmin
		newImageURL := "https://example.com/avatar.jpg"
		emailVerified := true
		newPasswordHash := "new_hashed_password"

		updateParams := types.UpdateUserParams{
			ID:            s.testUser.ID,
			Username:      &newUsername,
			Role:          &newRole,
			ImageURL:      &newImageURL,
			EmailVerified: &emailVerified,
			PasswordHash:  &newPasswordHash,
		}

		result := s.storage.Update(s.ctx, updateParams)
		assert.True(t, result.IsOk(), "Failed to update user: %v", result.Error)

		updatedUser := result.Must()
		assert.Equal(t, newUsername, updatedUser.Username)
		assert.Equal(t, newRole, updatedUser.Role)
		if newImageURL != "" {
			require.NotNil(t, updatedUser.ImageURL)
			assert.Equal(t, newImageURL, *updatedUser.ImageURL)
		} else {
			assert.Nil(t, updatedUser.ImageURL)
		}
		assert.Equal(t, emailVerified, updatedUser.EmailVerified)
		if newPasswordHash != "" {
			require.NotNil(t, updatedUser.PasswordHash)
			assert.Equal(t, newPasswordHash, *updatedUser.PasswordHash)
		} else {
			assert.Nil(t, updatedUser.PasswordHash)
		}
		assert.True(t, updatedUser.UpdatedAt.After(s.testUser.UpdatedAt))
	})

	t.Run("Частичное обновление", func(t *testing.T) {
		getResult := s.storage.GetByID(s.ctx, s.testUser.ID)
		require.True(t, getResult.IsOk())
		currentUser := getResult.Must()
		newUsername := "partialupdate"
		updateParams := types.UpdateUserParams{
			ID:       s.testUser.ID,
			Username: &newUsername,
		}

		result := s.storage.Update(s.ctx, updateParams)
		assert.True(t, result.IsOk())

		updatedUser := result.Must()
		assert.Equal(t, newUsername, updatedUser.Username)
		assert.Equal(t, currentUser.Role, updatedUser.Role) // Роль не изменилась
		if currentUser.ImageURL == nil {
			assert.Nil(t, updatedUser.ImageURL)
		} else {
			require.NotNil(t, updatedUser.ImageURL)
			assert.Equal(t, *currentUser.ImageURL, *updatedUser.ImageURL)
		}
	})

	t.Run("Обновление несуществующего пользователя", func(t *testing.T) {
		nonExistentID := uuid.New()
		newUsername := "nonexistent"

		updateParams := types.UpdateUserParams{
			ID:       nonExistentID,
			Username: &newUsername,
		}

		result := s.storage.Update(s.ctx, updateParams)
		assert.True(t, result.IsErr(), "Expected error for non-existent user")
	})

	t.Run("Обновление с дубликатом username", func(t *testing.T) {
		// Создаем второго пользователя
		secondUserParams := types.CreateUserParams{
			Email:    "second@example.com",
			Username: "seconduser",
			Role:     types.RoleCustomer,
		}

		secondUserResult := s.storage.Create(s.ctx, secondUserParams)
		require.True(t, secondUserResult.IsOk())

		// Пытаемся обновить первого пользователя с username второго
		duplicateUsername := "seconduser"
		updateParams := types.UpdateUserParams{
			ID:       s.testUser.ID,
			Username: &duplicateUsername,
		}

		result := s.storage.Update(s.ctx, updateParams)
		assert.True(t, result.IsOk(), "Should allow duplicate usernames")
	})

	t.Run("Обновление удаленного пользователя", func(t *testing.T) {
		// Создаем и удаляем пользователя
		createParams := types.CreateUserParams{
			Email:    "todeleteupdate@example.com",
			Username: "todeleteupdate",
			Role:     types.RoleCustomer,
		}

		createResult := s.storage.Create(s.ctx, createParams)
		require.True(t, createResult.IsOk())

		userToDelete := createResult.Must()

		// Удаляем
		err := s.storage.Delete(s.ctx, userToDelete.ID)
		require.NoError(t, err)

		// Пытаемся обновить
		newUsername := "updatedafterdelete"
		updateParams := types.UpdateUserParams{
			ID:       userToDelete.ID,
			Username: &newUsername,
		}

		result := s.storage.Update(s.ctx, updateParams)
		assert.True(t, result.IsErr(), "Expected error for deleted user")
	})
}

// TestList тестирует получение списка пользователей с пагинацией
func (s *UsersStorageTestSuite) TestList() {
	t := s.T()

	// Создаем несколько пользователей для тестирования
	users := []types.CreateUserParams{
		{
			Email:    "alice@example.com",
			Username: "alice",
			Role:     types.RoleCustomer,
		},
		{
			Email:    "bob@example.com",
			Username: "bob",
			Role:     types.RoleAdmin,
		},
		{
			Email:    "charlie@example.com",
			Username: "charlie",
			Role:     types.RoleEmployee,
		},
		{
			Email:    "david@example.com",
			Username: "david",
			Role:     types.RoleCustomer,
		},
		{
			Email:    "eve@example.com",
			Username: "eve",
			Role:     types.RoleCustomer,
		},
	}

	for _, userParams := range users {
		result := s.storage.Create(s.ctx, userParams)
		require.True(t, result.IsOk())
	}

	t.Run("Пагинация по умолчанию", func(t *testing.T) {
		params := types.ListUsersParams{
			Limit:          2,
			Offset:         0,
			IncludeDeleted: false,
		}

		result := s.storage.List(s.ctx, params)
		assert.True(t, result.IsOk(), "Failed to list users: %v", result.Error)

		response := result.Must()
		assert.Len(t, response.Data, 2)
		assert.Equal(t, params.Limit, response.Limit)
		assert.Equal(t, params.Offset, response.Offset)
		assert.True(t, response.Total >= 6) // 5 созданных + 1 тестовый
		assert.True(t, response.HasNextPage)
		assert.False(t, response.HasPreviousPage)
	})

	t.Run("Фильтрация по роли", func(t *testing.T) {
		role := types.RoleCustomer
		params := types.ListUsersParams{
			Limit:          10,
			Offset:         0,
			Role:           &role,
			IncludeDeleted: false,
		}

		result := s.storage.List(s.ctx, params)
		assert.True(t, result.IsOk())

		response := result.Must()

		// Проверяем, что все пользователи имеют указанную роль
		for _, user := range response.Data {
			assert.Equal(t, role, user.Role)
		}
	})

	t.Run("Поиск по email", func(t *testing.T) {
		email := "alice"
		params := types.ListUsersParams{
			Limit:          10,
			Offset:         0,
			Email:          &email,
			IncludeDeleted: false,
		}

		result := s.storage.List(s.ctx, params)
		assert.True(t, result.IsOk())

		response := result.Must()
		assert.GreaterOrEqual(t, len(response.Data), 1)

		for _, user := range response.Data {
			assert.Contains(t, strings.ToLower(user.Email), email)
		}
	})

	t.Run("Поиск по username", func(t *testing.T) {
		username := "bob"
		params := types.ListUsersParams{
			Limit:          10,
			Offset:         0,
			Username:       &username,
			IncludeDeleted: false,
		}

		result := s.storage.List(s.ctx, params)
		assert.True(t, result.IsOk())

		response := result.Must()
		assert.GreaterOrEqual(t, len(response.Data), 1)

		for _, user := range response.Data {
			assert.Contains(t, strings.ToLower(user.Username), username)
		}
	})

	t.Run("Общий поиск", func(t *testing.T) {
		search := "example"
		params := types.ListUsersParams{
			Limit:          10,
			Offset:         0,
			SearchQuery:    &search,
			IncludeDeleted: false,
		}

		result := s.storage.List(s.ctx, params)
		assert.True(t, result.IsOk())

		response := result.Must()
		assert.Greater(t, len(response.Data), 0)
	})

	t.Run("Сортировка по email", func(t *testing.T) {
		params := types.ListUsersParams{
			Limit:          10,
			Offset:         0,
			OrderBy:        "email",
			Order:          "ASC",
			IncludeDeleted: false,
		}

		result := s.storage.List(s.ctx, params)
		assert.True(t, result.IsOk())

		response := result.Must()
		assert.Greater(t, len(response.Data), 1)

		// Проверяем сортировку
		for i := 0; i < len(response.Data)-1; i++ {
			assert.True(t, response.Data[i].Email <= response.Data[i+1].Email)
		}
	})

	t.Run("Сортировка по дате создания", func(t *testing.T) {
		params := types.ListUsersParams{
			Limit:          10,
			Offset:         0,
			OrderBy:        "created_at",
			Order:          "DESC",
			IncludeDeleted: false,
		}

		result := s.storage.List(s.ctx, params)
		assert.True(t, result.IsOk())

		response := result.Must()
		assert.Greater(t, len(response.Data), 1)

		// Проверяем сортировку (последние созданные первыми)
		for i := 0; i < len(response.Data)-1; i++ {
			assert.True(t, response.Data[i].CreatedAt.After(response.Data[i+1].CreatedAt) ||
				response.Data[i].CreatedAt.Equal(response.Data[i+1].CreatedAt))
		}
	})

	t.Run("Пагинация - вторая страница", func(t *testing.T) {
		params := types.ListUsersParams{
			Limit:          2,
			Offset:         2,
			IncludeDeleted: false,
		}

		result := s.storage.List(s.ctx, params)
		assert.True(t, result.IsOk())

		response := result.Must()
		assert.Len(t, response.Data, 2)
		assert.Equal(t, 2, response.Offset)
		assert.True(t, response.HasNextPage)
		assert.True(t, response.HasPreviousPage)
	})

	t.Run("Включая удаленных пользователей", func(t *testing.T) {
		// Удаляем одного пользователя
		result := s.storage.GetByEmail(s.ctx, "alice@example.com")
		require.True(t, result.IsOk())
		alice := result.Must()

		err := s.storage.Delete(s.ctx, alice.ID)
		require.NoError(t, err)

		// Тест без OnlyActive (по умолчанию true)
		params1 := types.ListUsersParams{
			Limit:  100,
			Offset: 0,
		}
		result1 := s.storage.List(s.ctx, params1)
		assert.True(t, result1.IsOk())

		response1 := result1.Must()
		foundAlice1 := false
		for _, user := range response1.Data {
			if user.ID == alice.ID {
				foundAlice1 = true
				break
			}
		}
		assert.False(t, foundAlice1, "Deleted user should not appear when OnlyActive is true")

		// Тест с OnlyActive = false
		params2 := types.ListUsersParams{
			Limit:          100,
			Offset:         0,
			IncludeDeleted: true,
		}
		result2 := s.storage.List(s.ctx, params2)
		assert.True(t, result2.IsOk())

		response2 := result2.Must()
		foundAlice2 := false
		for _, user := range response2.Data {
			if user.ID == alice.ID {
				foundAlice2 = true
				break
			}
		}
		assert.True(t, foundAlice2, "Deleted user should appear when OnlyActive is false")
	})
}

// TestDelete тестирует мягкое удаление пользователя
func (s *UsersStorageTestSuite) TestDelete() {
	t := s.T()

	t.Run("Успешное удаление", func(t *testing.T) {
		// Создаем пользователя для удаления
		createParams := types.CreateUserParams{
			Email:    "todelete@example.com",
			Username: "todelete",
			Role:     types.RoleCustomer,
		}

		createResult := s.storage.Create(s.ctx, createParams)
		require.True(t, createResult.IsOk())

		userToDelete := createResult.Must()

		// Удаляем
		err := s.storage.Delete(s.ctx, userToDelete.ID)
		assert.NoError(t, err)

		// Проверяем, что пользователь не доступен через GetByID
		result := s.storage.GetByID(s.ctx, userToDelete.ID)
		assert.True(t, result.IsErr())

		// Проверяем, что пользователь есть в БД с установленным deleted_at
		var deletedAt *time.Time
		err = s.pool.QueryRow(s.ctx,
			"SELECT deleted_at FROM users WHERE id = $1",
			userToDelete.ID).Scan(&deletedAt)
		assert.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("Удаление несуществующего пользователя", func(t *testing.T) {
		nonExistentID := uuid.New()
		err := s.storage.Delete(s.ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Повторное удаление", func(t *testing.T) {
		// Создаем пользователя
		createParams := types.CreateUserParams{
			Email:    "doubledelete@example.com",
			Username: "doubledelete",
			Role:     types.RoleCustomer,
		}

		createResult := s.storage.Create(s.ctx, createParams)
		require.True(t, createResult.IsOk())

		user := createResult.Must()

		// Первое удаление
		err := s.storage.Delete(s.ctx, user.ID)
		assert.NoError(t, err)

		// Второе удаление
		err = s.storage.Delete(s.ctx, user.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

// TestConcurrentOperations тестирует конкурентные операции
func (s *UsersStorageTestSuite) TestConcurrentOperations() {
	t := s.T()

	const numGoroutines = 10
	errCh := make(chan error, numGoroutines)

	for i := range numGoroutines {
		go func(idx int) {
			email := fmt.Sprintf("concurrent%d@example.com", idx)
			username := fmt.Sprintf("concurrent%d", idx)

			// Создаем пользователя
			createParams := types.CreateUserParams{
				Email:    email,
				Username: username,
				Role:     types.RoleCustomer,
			}

			result := s.storage.Create(s.ctx, createParams)
			if result.IsErr() {
				errCh <- result.Error
				return
			}

			user := result.Must()

			// Получаем пользователя
			result = s.storage.GetByID(s.ctx, user.ID)
			if result.IsErr() {
				errCh <- result.Error
				return
			}

			// Обновляем пользователя
			newUsername := username + "_updated"
			updateParams := types.UpdateUserParams{
				ID:       user.ID,
				Username: &newUsername,
			}

			result = s.storage.Update(s.ctx, updateParams)
			if result.IsErr() {
				errCh <- result.Error
				return
			}

			errCh <- nil
		}(i)
	}

	// Собираем ошибки
	for i := 0; i < numGoroutines; i++ {
		err := <-errCh
		assert.NoError(t, err)
	}
}

// Вспомогательная функция для создания указателя на строку
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
