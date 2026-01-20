package types

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Email             string     `json:"email" db:"email"`
	EmailVerified     bool       `json:"email_verified" db:"email_verified"`
	VerificationToken *string    `json:"-" db:"verification_token"`
	Username          string     `json:"username" db:"username"`
	Role              UserRole   `json:"role" db:"role"`
	ImageURL          *string    `json:"image_url,omitempty" db:"image_url"`
	PasswordHash      *string    `json:"-" db:"password_hash"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PublicUser структура без чувствительных данных для публичного API
type PublicUser struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Username      string    `json:"username"`
	Role          UserRole  `json:"role"`
	ImageURL      string    `json:"image_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToPublic конвертирует User в PublicUser
func (u *User) ToPublic() PublicUser {
	pu := PublicUser{
		ID:            u.ID,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Username:      u.Username,
		Role:          u.Role,
		CreatedAt:     u.CreatedAt,
	}
	if u.ImageURL != nil {
		pu.ImageURL = *u.ImageURL
	}
	return pu
}

type CreateUserParams struct {
	Email             string
	Username          string
	PasswordHash      *string
	Role              UserRole
	ImageURL          *string
	VerificationToken *string
}

type UpdateUserParams struct {
	ID                uuid.UUID
	Username          *string
	Role              *UserRole
	ImageURL          *string
	EmailVerified     *bool
	VerificationToken *string
	PasswordHash      *string
}

type ListUsersParams struct {
	Limit          int
	Offset         int
	Role           *UserRole
	Email          *string
	Username       *string
	IncludeDeleted bool
	OrderBy        string
	Order          string // "ASC" или "DESC"
	SearchQuery    *string
}

func (p *ListUsersParams) BuildQuery() (query string, args pgx.NamedArgs) {
	var builder strings.Builder

	// Базовый SELECT
	builder.WriteString("SELECT * FROM users")

	// WHERE условия
	args = make(pgx.NamedArgs)
	conditions := []string{}

	// Обязательное условие - только активные пользователи
	if !p.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}
	// Фильтр по email
	if p.Email != nil && *p.Email != "" {
		conditions = append(conditions, "email ILIKE @email")
		args["email"] = "%" + *p.Email + "%"
	}

	// Фильтр по username
	if p.Username != nil && *p.Username != "" {
		conditions = append(conditions, "username ILIKE @username")
		args["username"] = "%" + *p.Username + "%"
	}

	// Фильтр по роли
	if p.Role != nil && *p.Role != "" {
		conditions = append(conditions, "role = @role")
		args["role"] = string(*p.Role)
	}

	// Общий поиск
	if p.SearchQuery != nil && *p.SearchQuery != "" {
		conditions = append(conditions, "(email ILIKE @search OR username ILIKE @search)")
		args["search"] = "%" + *p.SearchQuery + "%"
	}

	// Добавляем WHERE если есть условия
	if len(conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}

	// ORDER BY
	if p.OrderBy != "" {
		// Безопасный порядок сортировки
		safeOrderBy := "created_at"
		switch p.OrderBy {
		case "email", "username", "created_at", "updated_at", "role":
			safeOrderBy = p.OrderBy
		}

		builder.WriteString(" ORDER BY ")
		builder.WriteString(safeOrderBy)

		// Направление сортировки
		if strings.ToUpper(p.Order) == "ASC" {
			builder.WriteString(" ASC")
		} else {
			builder.WriteString(" DESC")
		}
	} else {
		// Сортировка по умолчанию
		builder.WriteString(" ORDER BY created_at DESC")
	}

	// LIMIT и OFFSET
	if p.Limit > 0 {
		builder.WriteString(" LIMIT @limit")
		args["limit"] = p.Limit
	}

	if p.Offset > 0 {
		builder.WriteString(" OFFSET @offset")
		args["offset"] = p.Offset
	}

	return builder.String(), args
}

// BuildCountQuery строит запрос для получения общего количества
func (p *ListUsersParams) BuildCountQuery() (query string, args pgx.NamedArgs) {
	var builder strings.Builder

	// Базовый SELECT COUNT
	builder.WriteString("SELECT COUNT(*) FROM users")

	// WHERE условия
	args = make(pgx.NamedArgs)
	conditions := []string{}

	// Обязательное условие - только активные пользователи
	if !p.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	// Фильтр по email
	if p.Email != nil && *p.Email != "" {
		conditions = append(conditions, "email ILIKE @email")
		args["email"] = "%" + *p.Email + "%"
	}

	// Фильтр по username
	if p.Username != nil && *p.Username != "" {
		conditions = append(conditions, "username ILIKE @username")
		args["username"] = "%" + *p.Username + "%"
	}

	// Фильтр по роли
	if p.Role != nil && *p.Role != "" {
		conditions = append(conditions, "role = @role")
		args["role"] = string(*p.Role)
	}

	// Общий поиск
	if p.SearchQuery != nil && *p.SearchQuery != "" {
		conditions = append(conditions, "(email ILIKE @search OR username ILIKE @search)")
		args["search"] = "%" + *p.SearchQuery + "%"
	}

	// Добавляем WHERE если есть условия
	if len(conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}

	return builder.String(), args
}
