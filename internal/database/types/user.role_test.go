package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRole_Valid(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		expected bool
	}{
		{"гость", RoleGuest, true},
		{"покупатель", RoleCustomer, true},
		{"сотрудник", RoleEmployee, true},
		{"администратор", RoleAdmin, true},
		{"владелец", RoleOwner, true},
		{"неизвестная роль", UserRole("unknown"), false},
		{"пустая роль", UserRole(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.Valid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserRole_String(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		expected string
	}{
		{"гость", RoleGuest, "guest"},
		{"покупатель", RoleCustomer, "customer"},
		{"сотрудник", RoleEmployee, "employee"},
		{"администратор", RoleAdmin, "admin"},
		{"владелец", RoleOwner, "owner"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserRole_MarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		role        UserRole
		expected    string
		expectError bool
	}{
		{"гость", RoleGuest, `"guest"`, false},
		{"покупатель", RoleCustomer, `"customer"`, false},
		{"сотрудник", RoleEmployee, `"employee"`, false},
		{"администратор", RoleAdmin, `"admin"`, false},
		{"владелец", RoleOwner, `"owner"`, false},
		{"некорректная роль", UserRole("invalid"), "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.role.MarshalJSON()

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestUserRole_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    UserRole
		expectError bool
	}{
		{"гость (нижний регистр)", `"guest"`, RoleGuest, false},
		{"покупатель (верхний регистр)", `"CUSTOMER"`, RoleCustomer, false},
		{"сотрудник (смешанный регистр)", `"EmPlOyEe"`, RoleEmployee, false},
		{"администратор", `"admin"`, RoleAdmin, false},
		{"владелец", `"owner"`, RoleOwner, false},
		{"пустая строка", `""`, RoleGuest, true},
		{"некорректная роль", `"superuser"`, RoleGuest, true},
		{"некорректный JSON", `invalid`, RoleGuest, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var role UserRole
			err := role.UnmarshalJSON([]byte(tt.input))

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, role)
		})
	}
}

func TestUserRole_getHierarchyLevel(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		expected int
	}{
		{"гость", RoleGuest, 0},
		{"покупатель", RoleCustomer, 1},
		{"сотрудник", RoleEmployee, 2},
		{"администратор", RoleAdmin, 3},
		{"владелец", RoleOwner, 4},
		{"неизвестная роль", UserRole("unknown"), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.role.getHierarchyLevel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserRole_HasPermission(t *testing.T) {
	tests := []struct {
		name     string
		userRole UserRole
		minLevel UserRole
		expected bool
	}{
		{
			name:     "администратор имеет доступ к уровню сотрудника",
			userRole: RoleAdmin,
			minLevel: RoleEmployee,
			expected: true,
		},
		{
			name:     "сотрудник не имеет доступа к уровню администратора",
			userRole: RoleEmployee,
			minLevel: RoleAdmin,
			expected: false,
		},
		{
			name:     "равные уровни",
			userRole: RoleCustomer,
			minLevel: RoleCustomer,
			expected: true,
		},
		{
			name:     "гость не имеет доступа к покупателю",
			userRole: RoleGuest,
			minLevel: RoleCustomer,
			expected: false,
		},
		{
			name:     "владелец имеет доступ к любому уровню",
			userRole: RoleOwner,
			minLevel: RoleGuest,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.userRole.HasPermission(tt.minLevel)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllRoles(t *testing.T) {
	roles := AllRoles()

	expected := []UserRole{
		RoleGuest,
		RoleCustomer,
		RoleEmployee,
		RoleAdmin,
		RoleOwner,
	}

	assert.Equal(t, expected, roles)
	assert.Len(t, roles, 5)

	// Проверяем, что все роли валидны
	for _, role := range roles {
		assert.True(t, role.Valid())
	}
}

func TestRoleFromString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    UserRole
		expectError bool
	}{
		{"гость (нижний регистр)", "guest", RoleGuest, false},
		{"покупатель (верхний регистр)", "CUSTOMER", RoleCustomer, false},
		{"сотрудник (смешанный регистр)", "EmPlOyEe", RoleEmployee, false},
		{"администратор", "admin", RoleAdmin, false},
		{"владелец", "owner", RoleOwner, false},
		{"пустая строка", "", RoleGuest, true},
		{"некорректная роль", "superuser", RoleGuest, true},
		{"строка с пробелами", " admin ", RoleGuest, true}, // Функция не делает trim
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := RoleFromString(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, RoleGuest, role)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, role)
		})
	}
}

// Тест для проверки совместимости с json.Marshal/Unmarshal
func TestUserRole_JSONCompatibility(t *testing.T) {
	type TestStruct struct {
		Role UserRole `json:"role"`
	}

	tests := []struct {
		name     string
		input    TestStruct
		expected string
	}{
		{
			name: "сериализация роли администратора",
			input: TestStruct{
				Role: RoleAdmin,
			},
			expected: `{"role":"admin"}`,
		},
		{
			name: "сериализация роли гостя",
			input: TestStruct{
				Role: RoleGuest,
			},
			expected: `{"role":"guest"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Тест маршалинга
			data, err := json.Marshal(tt.input)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))

			// Тест анмаршалинга
			var result TestStruct
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.input.Role, result.Role)
		})
	}

	// Тест на ошибку при анмаршалинге некорректной роли
	t.Run("ошибка при анмаршалинге некорректной роли", func(t *testing.T) {
		var result TestStruct
		err := json.Unmarshal([]byte(`{"role":"invalid"}`), &result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incorrect user role")
	})
}
