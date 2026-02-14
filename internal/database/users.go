package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LigeronAhill/luxcarpets-go/internal/database/types"
	"github.com/LigeronAhill/luxcarpets-go/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PgxPoolIface определяет интерфейс для работы с PostgreSQL пулом
// Это позволит использовать как реальный pgxpool.Pool, так и мок
type PgxPoolIface interface {
	Close()
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	Ping(ctx context.Context) error
}

type UsersStorage struct {
	pool PgxPoolIface
}

func NewUsersStorage(pool PgxPoolIface) *UsersStorage {
	return &UsersStorage{
		pool: pool,
	}
}

func (u *UsersStorage) Create(ctx context.Context, params types.CreateUserParams) (*types.User, error) {
	op := fmt.Sprintf("create new user\nparams:%#v", params)
	query := `
		INSERT INTO users (
		    email,
		    username,
		    password_hash,
		    role,
		    image_url,
		    verification_token
		)
		VALUES (@email, @username, @password_hash, @role, @image_url, @verification_token)
		RETURNING *
	`
	args := pgx.NamedArgs{
		"email":              strings.ToLower(params.Email),
		"username":           params.Username,
		"password_hash":      params.PasswordHash,
		"role":               params.Role,
		"image_url":          params.ImageURL,
		"verification_token": params.VerificationToken,
	}
	rows, err := u.pool.Query(ctx, query, args)
	if err != nil {
		if IsUniqueConstraintViolation(err, "users_email_key") {
			return nil, errors.New("email already exists")
		}
		return nil, utils.Wrap(op, err)
	}
	defer rows.Close()
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.User])
	if err != nil {
		if IsUniqueConstraintViolation(err, "users_email_key") {
			return nil, errors.New("email already exists")
		}
		return nil, utils.Wrap(op, err)
	}
	return res, nil
}

func (u *UsersStorage) GetByID(ctx context.Context, id uuid.UUID) (*types.User, error) {
	op := "get user by id " + id.String()
	query := `
		SELECT * FROM users WHERE id = @id AND deleted_at IS NULL
	`
	args := pgx.NamedArgs{
		"id": id,
	}
	rows, err := u.pool.Query(ctx, query, args)
	if err != nil {
		return nil, utils.Wrap(op, err)
	}
	defer rows.Close()
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.User])
	if err != nil {
		return nil, utils.Wrap(op, err)
	}
	return res, nil
}

func (u *UsersStorage) GetByEmail(ctx context.Context, email string) (*types.User, error) {
	op := "get user by email " + email
	query := `
		SELECT * FROM users WHERE email = @email AND deleted_at IS NULL
	`
	args := pgx.NamedArgs{
		"email": strings.ToLower(email),
	}
	rows, err := u.pool.Query(ctx, query, args)
	if err != nil {
		return nil, utils.Wrap(op, err)
	}
	defer rows.Close()
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.User])
	if err != nil {
		return nil, utils.Wrap(op, err)
	}
	return res, nil
}

func (u *UsersStorage) Update(ctx context.Context, params types.UpdateUserParams) (*types.User, error) {
	op := fmt.Sprintf("update user\nparams:%#v", params)
	query := `
		UPDATE users
		SET
		    username = COALESCE(@username, username),
		    role = COALESCE(@role, role),
		    image_url = COALESCE(@image_url, image_url),
		    email_verified = COALESCE(@email_verified, email_verified),
		    verification_token = COALESCE(@verification_token, verification_token),
		    password_hash = COALESCE(@password_hash, password_hash)
		WHERE id = @id AND deleted_at IS NULL
		RETURNING *;
	`
	args := pgx.NamedArgs{
		"id":                 params.ID,
		"username":           params.Username,
		"email_verified":     params.EmailVerified,
		"password_hash":      params.PasswordHash,
		"role":               params.Role,
		"image_url":          params.ImageURL,
		"verification_token": params.VerificationToken,
	}
	rows, err := u.pool.Query(ctx, query, args)
	if err != nil {
		return nil, utils.Wrap(op, err)
	}
	defer rows.Close()
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.User])
	if err != nil {
		return nil, utils.Wrap(op, err)
	}
	return res, nil
}

func (u *UsersStorage) List(ctx context.Context, params types.ListUsersParams) (*PaginatedResponse[*types.User], error) {
	op := fmt.Sprintf("list users\nparams:%#v", params)
	countQuery, countArgs := params.BuildCountQuery()
	var total int
	if err := u.pool.QueryRow(ctx, countQuery, countArgs).Scan(&total); err != nil {
		return nil, utils.Wrap(op, err)
	}
	query, args := params.BuildQuery()
	rows, err := u.pool.Query(ctx, query, args)
	if err != nil {
		return nil, utils.Wrap(op, err)
	}
	defer rows.Close()
	res, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[types.User])
	if err != nil {
		return nil, utils.Wrap(op, err)
	}
	return NewPaginatedResponse(res, total, params.Limit, params.Offset), nil
}

func (u *UsersStorage) Delete(ctx context.Context, id uuid.UUID) error {
	op := "delete user by id " + id.String()
	query := `
		UPDATE users SET deleted_at = NOW() WHERE id = @id AND deleted_at IS NULL;
	`
	args := pgx.NamedArgs{
		"id": id,
	}
	res, err := u.pool.Exec(ctx, query, args)
	if err != nil {
		return utils.Wrap(op, err)
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
