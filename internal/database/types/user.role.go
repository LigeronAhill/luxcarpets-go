package types

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UserRole представляет роль пользователя в системе
type UserRole string

const (
	RoleGuest    UserRole = "guest"
	RoleCustomer UserRole = "customer"
	RoleEmployee UserRole = "employee"
	RoleAdmin    UserRole = "admin"
	RoleOwner    UserRole = "owner"
)

// Valid проверяет, является ли роль допустимой
func (r UserRole) Valid() bool {
	switch r {
	case RoleGuest, RoleCustomer, RoleEmployee, RoleAdmin, RoleOwner:
		return true
	default:
		return false
	}
}

// String возвращает строковое представление роли
func (r UserRole) String() string {
	return string(r)
}

// MarshalJSON для сериализации в JSON
func (r UserRole) MarshalJSON() ([]byte, error) {
	if !r.Valid() {
		return nil, fmt.Errorf("incorrect user role: %s", r)
	}
	return json.Marshal(string(r))
}

// UnmarshalJSON для десериализации из JSON
func (r *UserRole) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	role := UserRole(strings.ToLower(s))
	if !role.Valid() {
		return fmt.Errorf("incorrect user role: %s", s)
	}
	*r = role
	return nil
}

// getHierarchyLevel возвращает уровень иерархии роли
func (r UserRole) getHierarchyLevel() int {
	switch r {
	case RoleGuest:
		return 0
	case RoleCustomer:
		return 1
	case RoleEmployee:
		return 2
	case RoleAdmin:
		return 3
	case RoleOwner:
		return 4
	default:
		return -1
	}
}

// HasPermission проверяет, имеет ли роль минимально необходимый уровень
func (r UserRole) HasPermission(minLevel UserRole) bool {
	return r.getHierarchyLevel() >= minLevel.getHierarchyLevel()
}

// AllRoles возвращает все допустимые роли
func AllRoles() []UserRole {
	return []UserRole{
		RoleGuest,
		RoleCustomer,
		RoleEmployee,
		RoleAdmin,
		RoleOwner,
	}
}

// RoleFromString создает UserRole из строки
func RoleFromString(s string) (UserRole, error) {
	role := UserRole(strings.ToLower(s))
	if !role.Valid() {
		return RoleGuest, fmt.Errorf("incorrect user role: %s", s)
	}
	return role, nil
}
