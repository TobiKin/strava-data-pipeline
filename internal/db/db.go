// filepath: /Users/kindtt01/Repositories/github/strava-data-pipeline/internal/db/db.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/TobiKin/strava-data-pipeline/internal/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// DB represents the database connection
type DB struct {
	*sql.DB
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

	db, err := sql.Open("postgres", connStr)
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
func (db *DB) InitSchema() error {
	// Read schema.sql and execute it
	schemaSQL := `
		-- Activities table
		CREATE TABLE IF NOT EXISTS activities (
			id BIGINT PRIMARY KEY,
			name TEXT,
			description TEXT,
			type TEXT,
			distance FLOAT,
			moving_time INT,
			elapsed_time INT,
			total_elevation_gain FLOAT,
			start_date TIMESTAMP,
			start_date_local TIMESTAMP,
			timezone TEXT,
			start_latlng TEXT,
			end_latlng TEXT,
			achievement_count INT,
			kudos_count INT,
			comment_count INT,
			athlete_count INT,
			photo_count INT,
			map_id TEXT,
			map_polyline TEXT,
			trainer BOOLEAN,
			commute BOOLEAN,
			manual BOOLEAN,
			private BOOLEAN,
			visibility TEXT,
			flagged BOOLEAN,
			workout_type INT,
			average_speed FLOAT,
			max_speed FLOAT,
			has_heartrate BOOLEAN,
			average_heartrate FLOAT,
			max_heartrate FLOAT,
			elev_high FLOAT,
			elev_low FLOAT,
			upload_id BIGINT,
			upload_id_str TEXT,
			external_id TEXT,
			athlete_id BIGINT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);

		-- Users table
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT PRIMARY KEY,
			username TEXT,
			firstname TEXT,
			lastname TEXT,
			city TEXT,
			country TEXT,
			sex TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			access_token TEXT,
			refresh_token TEXT,
			token_expires_at TIMESTAMP
		);

		-- API keys table
		CREATE TABLE IF NOT EXISTS api_keys (
			id SERIAL PRIMARY KEY,
			key TEXT UNIQUE NOT NULL,
			description TEXT,
			user_id BIGINT REFERENCES users(id),
			created_at TIMESTAMP DEFAULT NOW(),
			expires_at TIMESTAMP,
			is_active BOOLEAN DEFAULT TRUE
		);
	`
	_, err := db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("error creating schema: %w", err)
	}

	log.Println("Database schema initialized")
	return nil
}

// ValidateAPIKey checks if an API key is valid
func (db *DB) ValidateAPIKey(key string) (bool, error) {
	var isActive bool
	var expiresAt sql.NullTime

	query := `
		SELECT is_active, expires_at
		FROM api_keys
		WHERE key = $1
	`

	err := db.QueryRow(query, key).Scan(&isActive, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Key not found
		}
		return false, fmt.Errorf("error validating API key: %w", err)
	}

	// Check if key is active and not expired
	if !isActive {
		return false, nil
	}

	// If expiresAt is set, check if it's in the future
	if expiresAt.Valid {
		if expiresAt.Time.Before(time.Now()) {
			return false, nil
		}
	}

	return true, nil
}

// CreateAPIKey creates a new API key
func (db *DB) CreateAPIKey(key, description string, expiresAt *string) error {
	query := `
		INSERT INTO api_keys (key, description, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := db.Exec(query, key, description, expiresAt)
	if err != nil {
		return fmt.Errorf("error creating API key: %w", err)
	}

	return nil
}

// rowsToMaps converts SQL rows to a slice of maps
func rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		// Scan the result into the pointers
		if err := rows.Scan(pointers...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			rowMap[colName] = values[i]
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}
