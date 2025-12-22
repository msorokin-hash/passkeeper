package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/msorokin-hash/passkeeper/internal/storage"
	"github.com/msorokin-hash/passkeeper/internal/storage/postgres/migrations"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

// PGStorage represents a PostgreSQL-backed storage implementation.
// It manages a connection pool and provides access to the underlying database.
type PGStorage struct {
	pool *pgxpool.Pool
	log  *logrus.Entry
}

// NewPGStorage initializes a new PostgreSQL storage instance.
// It creates a connection pool, runs database migrations,
// and returns a Storage interface implementation.
// Returns a PGStorage instance or an error if initialization fails.
func NewPGStorage(ctx context.Context, dsn string, logger *logrus.Logger) (storage.Storage, error) {
	log := logger.WithField("instance", "pgStorage")

	pool, err := createPostgresPool(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err = createDabaseMigration(dsn, log); err != nil {
		return nil, err
	}

	return &PGStorage{pool: pool, log: log}, nil
}

// Close shuts down the PostgreSQL connection pool.
func (p *PGStorage) Close() error {
	p.pool.Close()
	return nil
}

// createPostgresPool creates and initializes a pgx connection pool.
// It validates the DSN, opens the pool, and pings the database.
// Returns a ready-to-use pool or an error if the connection cannot be established.
func createPostgresPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}

	return pool, nil
}

// createDabaseMigration runs database migrations using Goose.
// It configures the migration filesystem, logging, and dialect,
// opens a temporary DB connection, and applies migrations.
// Returns an error if migration setup or execution fails.
func createDabaseMigration(dsn string, logger *logrus.Entry) error {
	goose.SetBaseFS(migrations.FS)
	goose.SetLogger(logger)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := db.Close(); cerr != nil {
			logger.WithField("err", cerr).Error("failed to close db connection while migration")
		}
	}()

	if err = goose.Up(db, "."); err != nil {
		return err
	}

	return nil
}
