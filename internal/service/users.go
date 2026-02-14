package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LigeronAhill/luxcarpets-go/internal/database"
	"github.com/LigeronAhill/luxcarpets-go/internal/database/types"
	"github.com/google/uuid"
)

// Определяем переменные для ошибок
var (
	// Ошибки валидации
	// ErrEmailRequired возвращается, когда email не предоставлен
	ErrEmailRequired = errors.New("email cannot be empty")
	// ErrPasswordOrTokenReq возвращается, когда не предоставлен ни пароль, ни токен верификации
	ErrPasswordOrTokenReq = errors.New("either password or verification token must be provided")

	// Ошибки аутентификации
	// ErrWrongCredentials возвращается при неверных учетных данных
	ErrWrongCredentials = errors.New("wrong credentials")
	// ErrUserIDRequired возвращается при неверном ID пользователя
	ErrUserIDRequired = errors.New("user ID is required")
	// ErrUserNotFound возвращается если пользователь не найден
	ErrUserNotFound = errors.New("user not found")

	// Ошибки доступа
	// ErrPasswordLoginNotAvailable возвращается, когда у пользователя нет пароля
	ErrPasswordLoginNotAvailable = errors.New("password login not available for this user")
	// ErrTokenLoginNotAvailable возвращается, когда у пользователя нет токена верификации
	ErrTokenLoginNotAvailable = errors.New("token login not available for this user")

	// Ошибки пагинации
	ErrInvalidOffset = errors.New("offset must be greater than or equal to 0")
	ErrInvalidLimit  = errors.New("limit must be between 1 and 100")

	// Ошибки сортировки
	ErrInvalidOrderDirection = errors.New("order direction must be ASC or DESC")
)

// UsersService предоставляет методы для работы с пользователями
type UsersService struct {
	storage *database.UsersStorage
}

// NewUsersService создает новый экземпляр сервиса пользователей
func NewUsersService(storage *database.UsersStorage) *UsersService {
	return &UsersService{storage}
}

// SignUp регистрирует нового пользователя в системе
//
// Параметры:
//   - ctx: контекст выполнения
//   - email: email пользователя (обязательный)
//   - username: имя пользователя (обязательный)
//   - password: указатель на строку с паролем (может быть nil для OAuth регистрации)
//   - role: указатель на строку с ролью (может быть nil, тогда устанавливается RoleGuest)
//   - imageURL: указатель на строку с URL аватара (может быть nil)
//   - verificationToken: указатель на строку с токеном верификации (может быть nil)
//
// Возвращает:
//   - *types.PublicUser: публичные данные созданного пользователя
//   - error: ошибка, если регистрация не удалась
//
// Возможные ошибки:
//   - ошибки валидации пароля из функции hashPassword
//   - ErrInvalidRole если роль не существует
//   - ошибки базы данных при создании пользователя
func (s *UsersService) SignUp(ctx context.Context, email, username string, password, role, imageURL, verificationToken *string) (*types.PublicUser, error) {
	var passwordHash *string
	if password != nil {
		hash, err := hashPassword(*password)
		if err != nil {
			return nil, fmt.Errorf("invalid password: %w", err)
		}
		passwordHash = &hash
	}

	parsedRole := types.RoleGuest
	if role != nil {
		inputRole, err := types.RoleFromString(*role)
		if err != nil {
			return nil, fmt.Errorf("invalid role: %w", err)
		}
		parsedRole = inputRole
	}

	params := types.CreateUserParams{
		Email:             email,
		Username:          username,
		PasswordHash:      passwordHash,
		Role:              parsedRole,
		ImageURL:          imageURL,
		VerificationToken: verificationToken,
	}

	created, err := s.storage.Create(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	res := created.ToPublic()
	return &res, nil
}

// SignIn аутентифицирует пользователя по email и паролю или токену верификации
//
// Параметры:
//   - ctx: контекст выполнения
//   - email: email пользователя (обязательный)
//   - password: указатель на строку с паролем (может быть nil при входе по токену)
//   - verificationToken: указатель на строку с токеном верификации (может быть nil при входе по паролю)
//
// Возвращает:
//   - *types.PublicUser: публичные данные аутентифицированного пользователя
//   - error: ошибка, если аутентификация не удалась
//
// Возможные ошибки:
//   - ErrEmailRequired: если email не указан
//   - ErrPasswordOrTokenReq: если не указан ни пароль, ни токен
//   - ErrPasswordLoginNotAvailable: если у пользователя нет пароля (попытка входа по паролю)
//   - ErrTokenLoginNotAvailable: если у пользователя нет токена (попытка входа по токену)
//   - ErrWrongCredentials: если пароль или токен неверны
//   - ошибки базы данных при поиске пользователя
func (s *UsersService) SignIn(ctx context.Context, email string, password, verificationToken *string) (*types.PublicUser, error) {
	if email == "" {
		return nil, ErrEmailRequired
	}

	existing, err := s.storage.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	res := existing.ToPublic()

	// Вход по паролю
	if password != nil {
		// Проверяем, что у пользователя есть пароль
		if existing.PasswordHash == nil {
			return nil, ErrPasswordLoginNotAvailable
		}

		check, err := comparePasswordAndHash(*password, *existing.PasswordHash)
		if err != nil {
			return nil, fmt.Errorf("password verification failed: %w", err)
		}
		if !check {
			return nil, ErrWrongCredentials
		}
		return &res, nil
	}

	// Вход по токену верификации
	if verificationToken != nil {
		// Проверяем, что у пользователя есть токен
		if existing.VerificationToken == nil {
			return nil, ErrTokenLoginNotAvailable
		}

		// Сравниваем значения, а не указатели
		if *verificationToken != *existing.VerificationToken {
			return nil, ErrWrongCredentials
		}
		return &res, nil
	}

	// Если ни пароль, ни токен не предоставлены
	return nil, ErrPasswordOrTokenReq
}

// GetByID возвращает публичные данные пользователя по ID
//
// Параметры:
//   - ctx: контекст выполнения
//   - id: UUID пользователя
//
// Возвращает:
//   - *types.PublicUser: публичные данные пользователя
//   - error: ошибка, если пользователь не найден
//
// Возможные ошибки:
//   - ErrUserIDRequired: если ID не указан
//   - ErrUserNotFound: если пользователь не найден
//   - ошибки базы данных
func (s *UsersService) GetByID(ctx context.Context, id string) (*types.PublicUser, error) {
	if id == "" {
		return nil, ErrUserIDRequired
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid UUID format", ErrUserIDRequired)
	}

	user, err := s.storage.GetByID(ctx, parsedID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrUserNotFound, err)
	}

	res := user.ToPublic()
	return &res, nil
}

// List возвращает список пользователей с пагинацией, фильтрацией и сортировкой
//
// Параметры:
//   - ctx: контекст выполнения
//   - params: параметры для фильтрации, пагинации и сортировки
//
// Возвращает:
//   - *database.PaginatedResponse[*types.PublicUser]: список пользователей с информацией о пагинации
//   - error: ошибка, если получение списка не удалось
//
// Возможные ошибки:
//   - ErrInvalidOffset: если offset < 0
//   - ErrInvalidLimit: если limit < 1 или limit > 100
//   - ошибки базы данных
//
// Пример использования:
//
//	adminRole := types.RoleAdmin
//	searchQuery := "john"
//	params := types.ListUsersParams{
//	    Limit:       20,
//	    Offset:      0,
//	    Role:        &adminRole,
//	    SearchQuery: &searchQuery,
//	    OrderBy:     "created_at",
//	    Order:       "DESC",
//	}
//	response, err := service.List(ctx, params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Total users: %d\n", response.Total)
//	for _, user := range response.Items {
//	    fmt.Printf("- %s (%s) - %s\n", user.Username, user.Email, user.Role)
//	}
func (s *UsersService) List(ctx context.Context, params types.ListUsersParams) (*database.PaginatedResponse[*types.PublicUser], error) {
	// Валидация параметров пагинации
	if params.Offset < 0 {
		return nil, ErrInvalidOffset
	}
	if params.Limit < 1 || params.Limit > 100 {
		return nil, ErrInvalidLimit
	}

	// Дополнительная валидация, которую не делает BuildQuery
	if params.Order != "" {
		orderUpper := strings.ToUpper(params.Order)
		if orderUpper != "ASC" && orderUpper != "DESC" {
			return nil, ErrInvalidOrderDirection
		}
		params.Order = orderUpper
	}

	response, err := s.storage.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Преобразуем внутренних пользователей в публичные данные
	publicItems := make([]*types.PublicUser, len(response.Data))
	for i, user := range response.Data {
		publicUser := user.ToPublic()
		publicItems[i] = &publicUser
	}

	return database.NewPaginatedResponse(publicItems, response.Total, params.Limit, params.Offset), nil
}

// Delete мягко удаляет пользователя (устанавливает deleted_at)
//
// Параметры:
//   - ctx: контекст выполнения
//   - id: UUID пользователя для удаления
//
// Возвращает:
//   - error: ошибка, если удаление не удалось
//
// Возможные ошибки:
//   - ErrUserIDRequired: если ID не указан
//   - ErrUserNotFound: если пользователь не найден
//   - ошибки базы данных
//
// Примечание:
//   - Это "мягкое" удаление - запись помечается как удаленная, но остается в БД
//   - Удаленный пользователь не будет виден в списках и при поиске по ID,
//     если только в List не указан параметр IncludeDeleted = true
func (s *UsersService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrUserIDRequired
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%w: invalid UUID format", ErrUserIDRequired)
	}

	err = s.storage.Delete(ctx, parsedID)
	if err != nil {
		if err.Error() == "user not found" {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// Update обновляет данные пользователя
//
// Параметры:
//   - ctx: контекст выполнения
//   - params: параметры обновления (см. types.UpdateUserParams)
//
// Возвращает:
//   - *types.PublicUser: обновленные публичные данные пользователя
//   - error: ошибка, если обновление не удалось
//
// Возможные ошибки:
//   - ErrUserIDRequired: если ID не указан
//   - ErrUserNotFound: если пользователь не найден
//   - ошибки валидации пароля (если обновляется пароль)
//   - ошибки базы данных
//
// Пример использования:
//
//	newUsername := "new_username"
//	newImageURL := "https://example.com/avatar.jpg"
//	params := types.UpdateUserParams{
//	    ID:       userID,
//	    Username: &newUsername,
//	    ImageURL: &newImageURL,
//	}
//	updated, err := service.Update(ctx, params)
func (s *UsersService) Update(ctx context.Context, params types.UpdateUserParams) (*types.PublicUser, error) {
	if params.ID == uuid.Nil {
		return nil, ErrUserIDRequired
	}

	// Если обновляется пароль, проверяем его валидность
	if params.PasswordHash != nil {
		if err := validatePassword(*params.PasswordHash); err != nil {
			return nil, fmt.Errorf("invalid new password: %w", err)
		}

		// Хешируем новый пароль
		hash, err := hashPassword(*params.PasswordHash)
		if err != nil {
			return nil, fmt.Errorf("failed to hash new password: %w", err)
		}
		params.PasswordHash = &hash
	}

	updated, err := s.storage.Update(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	res := updated.ToPublic()
	return &res, nil
}

// GetByEmail возвращает публичные данные пользователя по email
//
// Параметры:
//   - ctx: контекст выполнения
//   - email: email пользователя
//
// Возвращает:
//   - *types.PublicUser: публичные данные пользователя
//   - error: ошибка, если пользователь не найден
//
// Возможные ошибки:
//   - ErrEmailRequired: если email не указан
//   - ErrUserNotFound: если пользователь не найден
//   - ошибки базы данных
func (s *UsersService) GetByEmail(ctx context.Context, email string) (*types.PublicUser, error) {
	if email == "" {
		return nil, ErrEmailRequired
	}

	user, err := s.storage.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrUserNotFound, err)
	}

	res := user.ToPublic()
	return &res, nil
}
