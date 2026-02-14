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
		return fmt.Errorf("error connecting to DB: %w", err)
	}
	defer conn.Close(ctx)

	fsys, err := fs.Sub(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("error getting file system: %w", err)
	}

	if _, err = conn.Exec(context.Background(), fmt.Sprintf("drop table if exists %s", versionTable)); err != nil {
		return fmt.Errorf("error cleaning versions table: %w", err)
	}
	m, err := migrate.NewMigrator(ctx, conn, versionTable)
	if err != nil {
		return fmt.Errorf("error creating migrator: %w", err)
	}

	if err = m.LoadMigrations(fsys); err != nil {
		return fmt.Errorf("error loading migrations: %w", err)
	}
	slog.Debug("migrations loaded", slog.Int("count", len(m.Migrations)))
	if err = m.Migrate(ctx); err != nil {
		return fmt.Errorf("error applying migrations: %w", err)
	}
	return nil
}
