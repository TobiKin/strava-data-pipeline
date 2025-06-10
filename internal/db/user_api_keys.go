package db

import (
	"fmt"
)

// AssociateAPIKeyWithUser associates an API key with a user
func (db *DB) AssociateAPIKeyWithUser(apiKey string, userID int64) error {
	query := `
		UPDATE api_keys
		SET user_id = $1
		WHERE key = $2
	`

	_, err := db.Exec(query, userID, apiKey)
	if err != nil {
		return fmt.Errorf("error associating API key with user: %w", err)
	}

	return nil
}

// GetAPIKeysForUser gets all API keys for a user
func (db *DB) GetAPIKeysForUser(userID int64) ([]map[string]interface{}, error) {
	query := `
		SELECT id, key, description, created_at, expires_at, is_active
		FROM api_keys
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying API keys: %w", err)
	}
	defer rows.Close()

	return rowsToMaps(rows)
}
