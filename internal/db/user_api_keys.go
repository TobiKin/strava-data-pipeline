package db

import (
	"fmt"
	"log"
	"time"
)

var apiKeySchema = `
CREATE TABLE IF NOT EXISTS api_keys (
	id BIGSERIAL PRIMARY KEY,
	key TEXT UNIQUE,
	description TEXT,
	created_at TIMESTAMP DEFAULT NOW(),
	expires_at TIMESTAMP,
	is_active BOOLEAN DEFAULT TRUE,
	user_id BIGINT
);`

type APIKey struct {
	ID          int64     `db:"id"`
	Key         string    `db:"key"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
	IsActive    bool      `db:"is_active"`
	UserID      *int64    `db:"user_id"`
}

// DB Schema for API keys
func (db *DB) CreateAPIKeySchema() {
	db.MustExec(apiKeySchema)
}

// ValidateAPIKey checks if an API key is valid
func (db *DB) ValidateAPIKey(key string) (bool, error) {
	var apiKey APIKey
	query := `
		SELECT is_active, expires_at
		FROM api_keys
		WHERE key = $1
	`
	err := db.Get(&apiKey, query, key)
	if err != nil {
		if isNoRows(err) {
			return false, nil // Key not found
		}
		return false, fmt.Errorf("error validating API key: %w", err)
	}
	if !apiKey.IsActive {
		return false, nil
	}
	if !apiKey.ExpiresAt.IsZero() && apiKey.ExpiresAt.Before(time.Now()) {
		return false, nil
	}
	return true, nil
}

/* -------------------------------------------------------------------------- */
/*                                CRUD API KEY                                */
/* -------------------------------------------------------------------------- */

// CreateAPIKey creates a new API key
func (db *DB) CreateAPIKey(key, description string, expiresAt *string) (APIKey, error) {
	var expiresAtTime time.Time
	if expiresAt != nil {
		var err error
		expiresAtTime, err = time.Parse(time.RFC3339, *expiresAt)
		if err != nil {
			return APIKey{}, fmt.Errorf("invalid expires_at format, expected RFC3339: %w", err)
		}
	}
	query := `
		INSERT INTO api_keys (key, description, expires_at)
		VALUES (:key, :description, :expires_at)
		RETURNING id, created_at, is_active
	`
	params := map[string]interface{}{
		"key":         key,
		"description": description,
		"expires_at":  expiresAtTime,
	}
	var apiKey APIKey
	err := db.QueryRowx(query, params["key"], params["description"], params["expires_at"]).Scan(&apiKey.ID, &apiKey.CreatedAt, &apiKey.IsActive)
	if err != nil {
		return APIKey{}, fmt.Errorf("error creating API key: %w", err)
	}
	apiKey.Key = key
	apiKey.Description = description
	apiKey.ExpiresAt = expiresAtTime
	return apiKey, nil
}

func (db *DB) ReadAPIKeyByID(id int64) (APIKey, error) {
	var apiKey APIKey
	query := `
		SELECT id, key, description, created_at, expires_at, is_active, user_id
		FROM api_keys
		WHERE id = $1
	`
	err := db.Get(&apiKey, query, id)
	if err != nil {
		if isNoRows(err) {
			return APIKey{}, fmt.Errorf("no API key found with the provided id %d", id)
		}
		return APIKey{}, fmt.Errorf("error reading API key: %w", err)
	}
	return apiKey, nil
}

func (db *DB) UpdateAPIKey(apiKey APIKey) (APIKey, error) {
	query := `
		UPDATE api_keys
		SET key = :key, description = :description, expires_at = :expires_at, is_active = :is_active, user_id = :user_id, updated_at = NOW()
		WHERE id = :id
		RETURNING created_at
	`
	params := map[string]interface{}{
		"id":          apiKey.ID,
		"key":         apiKey.Key,
		"description": apiKey.Description,
		"expires_at":  apiKey.ExpiresAt,
		"is_active":   apiKey.IsActive,
		"user_id":     apiKey.UserID,
	}
	var createdAt time.Time
	err := db.QueryRowx(query, params["key"], params["description"], params["expires_at"], params["is_active"], params["user_id"], params["id"]).Scan(&createdAt)
	if err != nil {
		return APIKey{}, fmt.Errorf("error updating API key: %w", err)
	}
	apiKey.CreatedAt = createdAt
	return apiKey, nil
}

func (db *DB) DeleteAPIKey(id int64) error {
	query := `
		DELETE FROM api_keys
		WHERE id = $1
	`
	result, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting API key: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no API key found with the provided id %d", id)
	} else {
		log.Printf("Deleted %d API key(s)", rowsAffected)
	}
	return nil
}

/* -------------------------------------------------------------------------- */
/*                                APIKEY + USER                               */
/* -------------------------------------------------------------------------- */

// AssociateAPIKeyWithUser associates an API key with a user
func (db *DB) AssociateAPIKeyWithUser(apiKey APIKey, userID int64) error {
	query := `
		UPDATE api_keys
		SET user_id = $1
		WHERE key = $2
	`
	_, err := db.Exec(query, userID, apiKey.Key)
	if err != nil {
		return fmt.Errorf("error associating API key with user: %w", err)
	}
	return nil
}

func (db *DB) ReadApiKeyByUserID(userID int64) ([]APIKey, error) {
	var apiKeys []APIKey
	query := `
		SELECT id, key, description, created_at, expires_at, is_active, user_id
		FROM api_keys
		WHERE user_id = $1
	`
	err := db.Select(&apiKeys, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error reading API keys for user %d: %w", userID, err)
	}
	return apiKeys, nil
}

// Helper to check for sqlx no rows error
func isNoRows(err error) bool {
	return err != nil && (err.Error() == "sql: no rows in result set" || err.Error() == "sqlx: no rows in result set")
}
