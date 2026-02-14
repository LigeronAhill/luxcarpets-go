package types

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestUser_ToPublic(t *testing.T) {
	tests := []struct {
		name     string
		user     *User
		expected PublicUser
	}{
		{
			name: "полный пользователь с ImageURL",
			user: &User{
				ID:            uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Email:         "test@example.com",
				EmailVerified: true,
				Username:      "testuser",
				Role:          RoleCustomer,
				ImageURL:      ptr("https://example.com/avatar.jpg"),
				PasswordHash:  ptr("hash123"),
				CreatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: PublicUser{
				ID:            uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Email:         "test@example.com",
				EmailVerified: true,
				Username:      "testuser",
				Role:          RoleCustomer,
				ImageURL:      "https://example.com/avatar.jpg",
				CreatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "пользователь без ImageURL",
			user: &User{
				ID:            uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Email:         "test@example.com",
				EmailVerified: false,
				Username:      "testuser",
				Role:          RoleAdmin,
				ImageURL:      nil,
				PasswordHash:  ptr("hash123"),
				CreatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			expected: PublicUser{
				ID:            uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
				Email:         "test@example.com",
				EmailVerified: false,
				Username:      "testuser",
				Role:          RoleAdmin,
				ImageURL:      "",
				CreatedAt:     time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.user.ToPublic()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListUsersParams_BuildQuery(t *testing.T) {
	tests := []struct {
		name          string
		params        *ListUsersParams
		expectedQuery string
		expectedArgs  pgx.NamedArgs
	}{
		{
			name:          "базовый запрос без параметров",
			params:        &ListUsersParams{},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC",
			expectedArgs:  pgx.NamedArgs{},
		},
		{
			name: "с включенными удаленными пользователями",
			params: &ListUsersParams{
				IncludeDeleted: true,
			},
			expectedQuery: "SELECT * FROM users ORDER BY created_at DESC",
			expectedArgs:  pgx.NamedArgs{},
		},
		{
			name: "с фильтром по email",
			params: &ListUsersParams{
				Email: ptr("test@example.com"),
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL AND email ILIKE @email ORDER BY created_at DESC",
			expectedArgs: pgx.NamedArgs{
				"email": "%test@example.com%",
			},
		},
		{
			name: "с фильтром по username",
			params: &ListUsersParams{
				Username: ptr("john"),
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL AND username ILIKE @username ORDER BY created_at DESC",
			expectedArgs: pgx.NamedArgs{
				"username": "%john%",
			},
		},
		{
			name: "с фильтром по роли",
			params: &ListUsersParams{
				Role: ptr(RoleAdmin),
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL AND role = @role ORDER BY created_at DESC",
			expectedArgs: pgx.NamedArgs{
				"role": string(RoleAdmin),
			},
		},
		{
			name: "с поисковым запросом",
			params: &ListUsersParams{
				SearchQuery: ptr("test"),
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL AND (email ILIKE @search OR username ILIKE @search) ORDER BY created_at DESC",
			expectedArgs: pgx.NamedArgs{
				"search": "%test%",
			},
		},
		{
			name: "с сортировкой по email ASC",
			params: &ListUsersParams{
				OrderBy: "email",
				Order:   "ASC",
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL ORDER BY email ASC",
			expectedArgs:  pgx.NamedArgs{},
		},
		{
			name: "с сортировкой по username DESC",
			params: &ListUsersParams{
				OrderBy: "username",
				Order:   "DESC",
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL ORDER BY username DESC",
			expectedArgs:  pgx.NamedArgs{},
		},
		{
			name: "с пагинацией",
			params: &ListUsersParams{
				Limit:  10,
				Offset: 20,
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT @limit OFFSET @offset",
			expectedArgs: pgx.NamedArgs{
				"limit":  10,
				"offset": 20,
			},
		},
		{
			name: "с комбинацией всех параметров",
			params: &ListUsersParams{
				Limit:       5,
				Offset:      10,
				Role:        ptr(RoleCustomer),
				Email:       ptr("test"),
				Username:    ptr("user"),
				OrderBy:     "created_at",
				Order:       "DESC",
				SearchQuery: ptr("query"),
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL AND email ILIKE @email AND username ILIKE @username AND role = @role AND (email ILIKE @search OR username ILIKE @search) ORDER BY created_at DESC LIMIT @limit OFFSET @offset",
			expectedArgs: pgx.NamedArgs{
				"email":    "%test%",
				"username": "%user%",
				"role":     string(RoleCustomer),
				"search":   "%query%",
				"limit":    5,
				"offset":   10,
			},
		},
		{
			name: "с небезопасным полем сортировки",
			params: &ListUsersParams{
				OrderBy: "password_hash",
				Order:   "ASC",
			},
			expectedQuery: "SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at ASC",
			expectedArgs:  pgx.NamedArgs{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := tt.params.BuildQuery()
			assert.Equal(t, tt.expectedQuery, query)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestListUsersParams_BuildCountQuery(t *testing.T) {
	tests := []struct {
		name          string
		params        *ListUsersParams
		expectedQuery string
		expectedArgs  pgx.NamedArgs
	}{
		{
			name:          "базовый подсчет",
			params:        &ListUsersParams{},
			expectedQuery: "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL",
			expectedArgs:  pgx.NamedArgs{},
		},
		{
			name: "подсчет с фильтром по роли",
			params: &ListUsersParams{
				Role: ptr(RoleEmployee),
			},
			expectedQuery: "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND role = @role",
			expectedArgs: pgx.NamedArgs{
				"role": string(RoleEmployee),
			},
		},
		{
			name: "подсчет с поиском",
			params: &ListUsersParams{
				SearchQuery: ptr("test"),
			},
			expectedQuery: "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL AND (email ILIKE @search OR username ILIKE @search)",
			expectedArgs: pgx.NamedArgs{
				"search": "%test%",
			},
		},
		{
			name: "подсчет с включенными удаленными",
			params: &ListUsersParams{
				IncludeDeleted: true,
			},
			expectedQuery: "SELECT COUNT(*) FROM users",
			expectedArgs:  pgx.NamedArgs{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := tt.params.BuildCountQuery()
			assert.Equal(t, tt.expectedQuery, query)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

// Вспомогательная функция для создания указателей
func ptr[T any](v T) *T {
	return &v
}
