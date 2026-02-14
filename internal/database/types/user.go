// Пакет types содержит основные структуры данных и параметры для работы с пользователями.
// Он определяет модели пользователей, параметры создания, обновления и фильтрации,
// а также методы для преобразования и построения запросов к базе данных.
package types

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// User представляет полную модель пользователя в системе.
// Содержит все поля, включая конфиденциальные данные (хэш пароля, токен верификации),
// которые не должны быть доступны в публичном API.
type User struct {
	ID                uuid.UUID  `json:"id" db:"id"`                           // Уникальный идентификатор пользователя
	Email             string     `json:"email" db:"email"`                     // Электронная почта пользователя
	EmailVerified     bool       `json:"email_verified" db:"email_verified"`   // Статус подтверждения email
	VerificationToken *string    `json:"-" db:"verification_token"`            // Токен для подтверждения email (не возвращается в JSON)
	Username          string     `json:"username" db:"username"`               // Имя пользователя
	Role              UserRole   `json:"role" db:"role"`                       // Роль пользователя в системе
	ImageURL          *string    `json:"image_url,omitempty" db:"image_url"`   // URL аватара пользователя (опционально)
	PasswordHash      *string    `json:"-" db:"password_hash"`                 // Хэш пароля (не возвращается в JSON)
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`           // Дата и время создания записи
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`           // Дата и время последнего обновления
	DeletedAt         *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // Дата мягкого удаления (nil = активная запись)
}

// PublicUser представляет публичную версию пользователя для API.
// Не содержит конфиденциальные данные, такие как хэш пароля или токены.
type PublicUser struct {
	ID            uuid.UUID `json:"id"`                  // Уникальный идентификатор пользователя
	Email         string    `json:"email"`               // Электронная почта пользователя
	EmailVerified bool      `json:"email_verified"`      // Статус подтверждения email
	Username      string    `json:"username"`            // Имя пользователя
	Role          UserRole  `json:"role"`                // Роль пользователя в системе
	ImageURL      string    `json:"image_url,omitempty"` // URL аватара пользователя
	CreatedAt     time.Time `json:"created_at"`          // Дата и время создания записи
}

// ToPublic преобразует полную модель пользователя в публичную версию.
// Возвращает PublicUser, безопасный для использования в публичных API.
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

// CreateUserParams содержит параметры для создания нового пользователя.
// Используется при регистрации или создании пользователя администратором.
type CreateUserParams struct {
	Email             string   // Электронная почта (обязательно)
	Username          string   // Имя пользователя (обязательно)
	PasswordHash      *string  // Хэш пароля (nil для OAuth пользователей)
	Role              UserRole // Роль пользователя (по умолчанию UserRoleUser)
	ImageURL          *string  // URL аватара (опционально)
	VerificationToken *string  // Токен для подтверждения email (опционально)
}

// UpdateUserParams содержит параметры для обновления существующего пользователя.
// Все поля опциональны - обновляются только переданные значения.
type UpdateUserParams struct {
	ID                uuid.UUID // ID пользователя для обновления (обязательно)
	Username          *string   // Новое имя пользователя
	Role              *UserRole // Новая роль
	ImageURL          *string   // Новый URL аватара
	EmailVerified     *bool     // Новый статус подтверждения email
	VerificationToken *string   // Новый токен верификации
	PasswordHash      *string   // Новый хэш пароля
}

// ListUsersParams содержит параметры фильтрации, пагинации и сортировки
// для получения списка пользователей.
type ListUsersParams struct {
	Limit          int       // Максимальное количество записей
	Offset         int       // Смещение для пагинации
	Role           *UserRole // Фильтр по роли
	Email          *string   // Поиск по email (частичное совпадение)
	Username       *string   // Поиск по имени (частичное совпадение)
	IncludeDeleted bool      // Включать ли мягко удаленных пользователей
	OrderBy        string    // Поле для сортировки (created_at, email, username, role, updated_at)
	Order          string    // Направление сортировки (ASC или DESC)
	SearchQuery    *string   // Глобальный поиск по email и username
}

// BuildQuery формирует SQL запрос для получения списка пользователей
// с учетом всех параметров фильтрации, сортировки и пагинации.
// Возвращает строку запроса и именованные аргументы для pgx.
func (p *ListUsersParams) BuildQuery() (query string, args pgx.NamedArgs) {
	var builder strings.Builder

	builder.WriteString("SELECT * FROM users")

	args = make(pgx.NamedArgs)
	conditions := []string{}

	// Исключаем мягко удаленных пользователей, если не указано обратное
	if !p.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	// Добавляем фильтры только для непустых значений
	if p.Email != nil && *p.Email != "" {
		conditions = append(conditions, "email ILIKE @email")
		args["email"] = "%" + *p.Email + "%"
	}

	if p.Username != nil && *p.Username != "" {
		conditions = append(conditions, "username ILIKE @username")
		args["username"] = "%" + *p.Username + "%"
	}

	if p.Role != nil && *p.Role != "" {
		conditions = append(conditions, "role = @role")
		args["role"] = string(*p.Role)
	}

	if p.SearchQuery != nil && *p.SearchQuery != "" {
		conditions = append(conditions, "(email ILIKE @search OR username ILIKE @search)")
		args["search"] = "%" + *p.SearchQuery + "%"
	}

	// Добавляем WHERE если есть условия
	if len(conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}

	// Добавляем сортировку (с проверкой безопасных полей)
	if p.OrderBy != "" {
		safeOrderBy := "created_at"
		switch p.OrderBy {
		case "email", "username", "created_at", "updated_at", "role":
			safeOrderBy = p.OrderBy
		}

		builder.WriteString(" ORDER BY ")
		builder.WriteString(safeOrderBy)

		if strings.ToUpper(p.Order) == "ASC" {
			builder.WriteString(" ASC")
		} else {
			builder.WriteString(" DESC")
		}
	} else {
		builder.WriteString(" ORDER BY created_at DESC")
	}

	// Добавляем LIMIT и OFFSET для пагинации
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

// BuildCountQuery формирует SQL запрос для подсчета общего количества
// пользователей, соответствующих критериям фильтрации (без пагинации).
// Используется для построения пагинации в API.
func (p *ListUsersParams) BuildCountQuery() (query string, args pgx.NamedArgs) {
	var builder strings.Builder

	builder.WriteString("SELECT COUNT(*) FROM users")

	args = make(pgx.NamedArgs)
	conditions := []string{}

	// Применяем те же фильтры, что и в BuildQuery
	if !p.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	if p.Email != nil && *p.Email != "" {
		conditions = append(conditions, "email ILIKE @email")
		args["email"] = "%" + *p.Email + "%"
	}

	if p.Username != nil && *p.Username != "" {
		conditions = append(conditions, "username ILIKE @username")
		args["username"] = "%" + *p.Username + "%"
	}

	if p.Role != nil && *p.Role != "" {
		conditions = append(conditions, "role = @role")
		args["role"] = string(*p.Role)
	}

	if p.SearchQuery != nil && *p.SearchQuery != "" {
		conditions = append(conditions, "(email ILIKE @search OR username ILIKE @search)")
		args["search"] = "%" + *p.SearchQuery + "%"
	}

	if len(conditions) > 0 {
		builder.WriteString(" WHERE ")
		builder.WriteString(strings.Join(conditions, " AND "))
	}

	return builder.String(), args
}
