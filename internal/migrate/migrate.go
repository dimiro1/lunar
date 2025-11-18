// Package migrate provides database migration functionality using golang-migrate.
// Migration files are embedded and tracked in a schema_migrations table.
package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var FS embed.FS

// Run executes all pending database migrations from the embedded filesystem.
// Only migrations that haven't been applied yet will run.
// Returns an error if the database is in a dirty state or if any migration fails.
func Run(db *sql.DB, migrationFS fs.FS) error {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d, please fix manually", version)
	}

	// Run migrations
	slog.Info("Running database migrations", "current_version", version)

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		slog.Info("No pending migrations to run")
	} else {
		newVersion, _, err := m.Version()
		if err != nil {
			return fmt.Errorf("failed to get new migration version: %w", err)
		}
		slog.Info("Migrations completed successfully", "new_version", newVersion)
	}

	return nil
}

// RunTest is a helper for running migrations in tests.
// It wraps Run and fails the test on error.
func RunTest(t interface {
	Helper()
	Fatalf(format string, args ...any)
}, db *sql.DB,
) {
	t.Helper()

	if err := Run(db, FS); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
}

// Status returns the current migration version and dirty state.
// A dirty state indicates a migration failed partway through and needs manual intervention.
func Status(db *sql.DB, migrationFS fs.FS) (version uint, dirty bool, err error) {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return 0, false, fmt.Errorf("failed to create source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration instance: %w", err)
	}

	version, dirty, err = m.Version()
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}

	return version, dirty, err
}
