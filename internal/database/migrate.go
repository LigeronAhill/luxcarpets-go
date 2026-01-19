package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

const versionTable = "schema_version"

func migrateDB(ctx context.Context, dbURL string) error {
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}
	defer conn.Close(ctx)

	fsys, err := fs.Sub(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("ошибка получения файловой системы: %w", err)
	}

	if _, err = conn.Exec(context.Background(), fmt.Sprintf("drop table if exists %s", versionTable)); err != nil {
		return fmt.Errorf("ошибка удаления таблицы версии схемы: %w", err)
	}
	m, err := migrate.NewMigrator(ctx, conn, versionTable)
	if err != nil {
		return fmt.Errorf("ошибка создания мигратора: %w", err)
	}

	if err = m.LoadMigrations(fsys); err != nil {
		return fmt.Errorf("ошибка загрузки миграций: %w", err)
	}
	slog.Debug("Загрузка миграций завершена", slog.Int("загружено", len(m.Migrations)))
	if err = m.Migrate(ctx); err != nil {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}
	return nil
}
