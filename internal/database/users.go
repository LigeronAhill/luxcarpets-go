package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/LigeronAhill/luxcarpets-go/internal/database/types"
	r "github.com/LigeronAhill/luxcarpets-go/pkg/result"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersStorage struct {
	pool *pgxpool.Pool
}

func NewUsersStorage(pool *pgxpool.Pool) *UsersStorage {
	return &UsersStorage{
		pool: pool,
	}
}

func (u *UsersStorage) Create(ctx context.Context, params types.CreateUserParams) r.Result[*types.User] {
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
	return r.AndThen(r.Try(u.pool.Query(ctx, query, args)), func(rows pgx.Rows) r.Result[*types.User] {
		defer rows.Close()
		return r.Try(pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.User])).MapErr(func(err error) error {
			if IsUniqueConstraintViolation(err, "users_email_key") {
				return errors.New("email already exists")
			}
			return err
		})
	})
}
func (u *UsersStorage) GetByID(ctx context.Context, id uuid.UUID) r.Result[*types.User] {
	query := `
		SELECT * FROM users WHERE id = @id AND deleted_at IS NULL
	`
	args := pgx.NamedArgs{
		"id": id,
	}
	return r.AndThen(r.Try(u.pool.Query(ctx, query, args)), func(rows pgx.Rows) r.Result[*types.User] {
		defer rows.Close()
		return r.Try(pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.User]))
	})
}
func (u *UsersStorage) GetByEmail(ctx context.Context, email string) r.Result[*types.User] {
	query := `
		SELECT * FROM users WHERE email = @email AND deleted_at IS NULL
	`
	args := pgx.NamedArgs{
		"email": strings.ToLower(email),
	}
	return r.AndThen(r.Try(u.pool.Query(ctx, query, args)), func(rows pgx.Rows) r.Result[*types.User] {
		defer rows.Close()
		return r.Try(pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.User]))
	})
}
func (u *UsersStorage) Update(ctx context.Context, params types.UpdateUserParams) r.Result[*types.User] {
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
	return r.AndThen(r.Try(u.pool.Query(ctx, query, args)), func(rows pgx.Rows) r.Result[*types.User] {
		defer rows.Close()
		return r.Try(pgx.CollectExactlyOneRow(rows, pgx.RowToAddrOfStructByName[types.User]))
	})
}
func (u *UsersStorage) List(ctx context.Context, params types.ListUsersParams) r.Result[PaginatedResponse[*types.User]] {
	countQuery, countArgs := params.BuildCountQuery()
	var total int
	if err := u.pool.QueryRow(ctx, countQuery, countArgs).Scan(&total); err != nil {
		return r.Err[PaginatedResponse[*types.User]](err)
	}
	query, args := params.BuildQuery()
	rows, err := u.pool.Query(ctx, query, args)
	if err != nil {
		return r.Err[PaginatedResponse[*types.User]](err)
	}
	defer rows.Close()
	res, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[types.User])
	if err != nil {
		return r.Err[PaginatedResponse[*types.User]](err)
	}
	return r.Ok(NewPaginatedResponse(res, total, params.Limit, params.Offset))
}
func (u *UsersStorage) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users SET deleted_at = NOW() WHERE id = @id AND deleted_at IS NULL;
	`
	args := pgx.NamedArgs{
		"id": id,
	}
	res, err := u.pool.Exec(ctx, query, args)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
