package db

import (
	"fmt"
	"time"
)

// UpdateUserTokens updates the OAuth tokens for a user
func (db *DB) UpdateUserTokens(userID int64, accessToken, refreshToken string, expiresAt time.Time) error {
	query := `
		UPDATE users
		SET access_token = $1, refresh_token = $2, token_expires_at = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := db.Exec(query, accessToken, refreshToken, expiresAt, userID)
	if err != nil {
		return fmt.Errorf("error updating user tokens: %w", err)
	}

	return nil
}

// GetUserTokens retrieves the OAuth tokens for a user
func (db *DB) GetUserTokens(userID int64) (accessToken, refreshToken string, expiresAt time.Time, err error) {
	query := `
		SELECT access_token, refresh_token, token_expires_at
		FROM users
		WHERE id = $1
	`

	err = db.QueryRow(query, userID).Scan(&accessToken, &refreshToken, &expiresAt)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("error retrieving user tokens: %w", err)
	}

	return accessToken, refreshToken, expiresAt, nil
}
