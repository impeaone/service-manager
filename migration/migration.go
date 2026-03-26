package migration

import (
	"ServiceManager/internal/config"
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.postgres.sql
var FSPG embed.FS

//go:embed *.sqlite.sql
var FSSQ embed.FS

const dirPath = "."

const postgresDB = "postgres"
const sqliteDB = "sqlite3"

var ErrorNoChange = fmt.Errorf("no change")

func Migrate(ctx context.Context) error {
	cfg, ok := ctx.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("config error")
	}
	switch cfg.DBType {
	case postgresDB:
		url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
		)

		db, err := sql.Open(postgresDB, url)
		if err != nil {
			return fmt.Errorf("connect postgres: %w", err)
		}
		defer func() { _ = db.Close() }()

		return runPostgresMigrations(db)
	case sqliteDB:
		db, err := sql.Open(sqliteDB, cfg.DBFilePath)
		if err != nil {
			return fmt.Errorf("connect sqlite3: %w", err)
		}
		defer func() { _ = db.Close() }()

		return runSqliteMigrations(db)
	default:
		return fmt.Errorf("unknown database type: %s", cfg.DBType)
	}
}

func runPostgresMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	sourceDriver, err := iofs.New(FSPG, dirPath)
	if err != nil {
		return fmt.Errorf("create source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, postgresDB, driver)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}

	if err = m.Up(); err != nil && errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

func runSqliteMigrations(db *sql.DB) error {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("create sqlite driver: %w", err)
	}

	sourceDriver, err := iofs.New(FSSQ, dirPath)
	if err != nil {
		return fmt.Errorf("create source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, sqliteDB, driver)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}

	if err = m.Up(); err != nil && errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil

}
