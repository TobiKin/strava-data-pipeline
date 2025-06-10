package db

import (
	"testing"
	"time"
)

func setupTestUserDB(t *testing.T) *DB {
	db := setupTestDB(t)
	db.CreateUserSchema()
	return db
}

func createTestUser(t *testing.T, db *DB) User {
	user := User{
		ID:        1,
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	query := `INSERT INTO User (id, username, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, user.ID, user.Username, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func TestCreateUser(t *testing.T) {
	db := setupTestUserDB(t)
	defer db.Close()
	_ = createTestUser(t, db)
}

func TestGetUserByID(t *testing.T) {
	db := setupTestUserDB(t)
	defer db.Close()
	created := createTestUser(t, db)
	var user User
	query := `SELECT * FROM User WHERE id = $1`
	err := db.Get(&user, query, created.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}
	if user.ID != created.ID {
		t.Fatalf("Expected user ID %d, got %d", created.ID, user.ID)
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestUserDB(t)
	defer db.Close()
	created := createTestUser(t, db)
	created.Username = "updateduser"
	query := `UPDATE User SET username = $1, updated_at = $2 WHERE id = $3`
	_, err := db.Exec(query, created.Username, time.Now(), created.ID)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	var user User
	err = db.Get(&user, `SELECT * FROM User WHERE id = $1`, created.ID)
	if err != nil {
		t.Fatalf("Failed to get user after update: %v", err)
	}
	if user.Username != "updateduser" {
		t.Fatalf("Expected updated username, got %s", user.Username)
	}
}

func TestDeleteUser(t *testing.T) {
	db := setupTestUserDB(t)
	defer db.Close()
	created := createTestUser(t, db)
	_, err := db.Exec(`DELETE FROM User WHERE id = $1`, created.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	var user User
	err = db.Get(&user, `SELECT * FROM User WHERE id = $1`, created.ID)
	if err == nil {
		t.Fatal("Expected error for deleted user, got nil")
	}
}
