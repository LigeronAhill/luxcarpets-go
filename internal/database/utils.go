package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsUniqueConstraintViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && (pgErr.ConstraintName == constraintName || constraintName == "") // Проверяем только код ошибки
	}
	return false
}
