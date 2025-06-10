package db

import (
	"fmt"
	"time"
)

var userSchema = `
CREATE TABLE IF NOT EXISTS users (
	id BIGINT PRIMARY KEY,
	username TEXT UNIQUE,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW(),
	access_token TEXT,
	refresh_token TEXT,
	token_expires_at TIMESTAMP
);`

type User struct {
	ID             int64     `db:"id"`
	Username       string    `db:"username"`
	AthleteID      int64     `db:"athlete_id"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	AccessToken    string    `db:"access_token"`
	RefreshToken   string    `db:"refresh_token"`
	TokenExpiresAt time.Time `db:"token_expires_at"`
}

func (db *DB) CreateUserSchema() {
	db.MustExec(userSchema)
}

func (db *DB) CreateUser(username string, athleteID int64) (User, error) {
	user := User{
		Username:  username,
		AthleteID: athleteID,
	}

	// Todo: Check if user already exists

	query := `
		INSERT INTO users (username, athlete_id)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := db.QueryRow(query, username, athleteID).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return User{}, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

func (db *DB) GetUserByID(userID int64) (User, error) {
	user := User{}

	query := `
		SELECT id, username, created_at, updated_at, access_token, refresh_token, token_expires_at
		FROM users
		WHERE id = $1
	`

	err := db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt,
		&user.AccessToken, &user.RefreshToken, &user.TokenExpiresAt)
	if err != nil {
		return User{}, fmt.Errorf("error retrieving user: %w", err)
	}

	return user, nil
}

func (db *DB) GetUserByUsername(username string) (User, error) {
	user := User{}

	query := `
		SELECT id, username, created_at, updated_at, access_token, refresh_token, token_expires_at
		FROM users
		WHERE username = $1
	`

	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt,
		&user.AccessToken, &user.RefreshToken, &user.TokenExpiresAt)
	if err != nil {
		return User{}, fmt.Errorf("error retrieving user by username: %w", err)
	}

	return user, nil
}

func (db *DB) GetUserByAthleteID(athleteID int64) (User, error) {
	user := User{}

	query := `
		SELECT id, username, created_at, updated_at, access_token, refresh_token, token_expires_at
		FROM users
		WHERE athlete_id = $1
	`

	err := db.QueryRow(query, athleteID).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt,
		&user.AccessToken, &user.RefreshToken, &user.TokenExpiresAt)
	if err != nil {
		return User{}, fmt.Errorf("error retrieving user by athlete ID: %w", err)
	}

	return user, nil
}

func (db *DB) UpdateUser(user User) error {
	query := `
		UPDATE users
		SET username = $1, athlete_id = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := db.Exec(query, user.Username, user.AthleteID, user.ID)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

func (db *DB) DeleteUser(userID int64) error {
	query := `
		DELETE FROM users
		WHERE id = $1
	`

	_, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}
