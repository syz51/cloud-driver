package database

import (
	"context"
	"fmt"
	"log"

	"cloud-driver/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Database wraps the database connection and provides database operations
type Database struct {
	pool *pgxpool.Pool
}

// New creates a new database instance with connection pool
func New(cfg *config.DatabaseConfig) (*Database, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection established successfully")

	return &Database{
		pool: pool,
	}, nil
}

// Close closes the database connection pool
func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
		log.Println("Database connection pool closed")
	}
}

// GetPool returns the underlying connection pool
func (db *Database) GetPool() *pgxpool.Pool {
	return db.pool
}

// WithTx executes a function within a database transaction
func (db *Database) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback(ctx)
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Printf("failed to rollback transaction: %v", rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
