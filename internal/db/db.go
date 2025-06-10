// filepath: /Users/kindtt01/Repositories/github/strava-data-pipeline/internal/db/db.go
package db

import (
	"fmt"

	"github.com/TobiKin/strava-data-pipeline/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB represents the database connection
type DB struct {
	*sqlx.DB
}

// New creates a new database connection
func New(config *config.Config) (*DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// InitSchema initializes the database schema
func (db *DB) InitSchema() {
	db.CreateUserSchema()
	db.CreateAPIKeySchema()
	db.CreateActivitySchema()
}
