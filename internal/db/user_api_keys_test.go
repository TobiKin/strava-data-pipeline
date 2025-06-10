package db

import (
	"testing"
	"time"

	"github.com/TobiKin/strava-data-pipeline/internal/config"
	"github.com/google/uuid"
)

const (
	DATABASE_HOST     = "192.168.64.5"
	DATABASE_PORT     = 5432
	DATABASE_USER     = "user"
	DATABASE_PASSWORD = "password"
	DATABASE_NAME     = "tempdb"
	DATABASE_SSLMODE  = "disable"
)

func setupTestConfig() *config.Config {
	return &config.Config{
		Database: config.Database{
			Host:     DATABASE_HOST,
			Port:     DATABASE_PORT,
			User:     DATABASE_USER,
			Password: DATABASE_PASSWORD,
			Name:     DATABASE_NAME,
			SSLMode:  DATABASE_SSLMODE,
		},
	}
}

func setupTestDB(t *testing.T) *DB {
	config := setupTestConfig()
	db, err := New(config)
	if err != nil {
		t.Fatalf("Failed to create new DB: %v", err)
	}

	if db == nil {
		t.Fatal("Expected non-nil DB instance")
	}

	return db
}

func TestNew(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	if db == nil {
		t.Fatal("Expected non-nil DB instance")
	}
}

func TestInitSchema(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	db.CreateAPIKeySchema()
	defer db.Close()
}

func createTestAPIKey(t *testing.T, db *DB) APIKey {
	expiresAt := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	key := "test_" + uuid.New().String()
	apiKey, err := db.CreateAPIKey(key, "Test API Key", &expiresAt)
	if err != nil {
		t.Fatalf("Failed to create test API key: %v", err)
	}
	return apiKey
}

func deleteTestAPIKey(t *testing.T, db *DB, apiKey APIKey) {
	if err := db.DeleteAPIKey(apiKey.ID); err != nil {
		t.Fatalf("Failed to delete test API key: %v", err)
	}
}

func TestValidateAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	apiKey := createTestAPIKey(t, db)
	defer deleteTestAPIKey(t, db, apiKey)

	if valid, err := db.ValidateAPIKey(apiKey.Key); err != nil {
		t.Fatalf("API key validation failed: %v", err)
	} else if !valid {
		t.Fatal("Expected valid API key")
	}
}

func TestReadApiKeyByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	apiKey := createTestAPIKey(t, db)
	defer deleteTestAPIKey(t, db, apiKey)

	retrievedKey, err := db.ReadAPIKeyByID(apiKey.ID)
	if err != nil {
		t.Fatalf("Failed to read API key by ID: %v", err)
	}

	if retrievedKey.ID != apiKey.ID {
		t.Fatalf("Expected API key ID %d, got %d", apiKey.ID, retrievedKey.ID)
	}
}

func TestUpdateAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	apiKey := createTestAPIKey(t, db)
	defer deleteTestAPIKey(t, db, apiKey)

	apiKey.Description = "Updated Test API Key"
	apiKey.ExpiresAt = time.Now().Add(2 * time.Hour).UTC()
	updatedKey, err := db.UpdateAPIKey(apiKey)
	if err != nil {
		t.Fatalf("Failed to update API key: %v", err)
	}
	if updatedKey.Description != apiKey.Description {
		t.Fatalf("Expected updated description %s, got %s", apiKey.Description, updatedKey.Description)
	}
	if updatedKey.ExpiresAt != apiKey.ExpiresAt {
		t.Fatalf("Expected updated expires_at %s, got %s", apiKey.ExpiresAt, updatedKey.ExpiresAt)
	}
	if updatedKey.Key != apiKey.Key {
		t.Fatalf("Expected updated key %s, got %s", apiKey.Key, updatedKey.Key)
	}
	if updatedKey.ID != apiKey.ID {
		t.Fatalf("Expected updated ID %d, got %d", apiKey.ID, updatedKey.ID)
	}
	if updatedKey.CreatedAt.IsZero() {
		t.Fatal("Expected non-zero CreatedAt timestamp")
	}
	if updatedKey.IsActive != apiKey.IsActive {
		t.Fatalf("Expected IsActive %v, got %v", apiKey.IsActive, updatedKey.IsActive)
	}
	if updatedKey.CreatedAt.Before(apiKey.CreatedAt) {
		t.Fatalf("Expected CreatedAt %s to be after original %s", updatedKey.CreatedAt, apiKey.CreatedAt)
	}
}

func TestDeleteAPIKey(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	apiKey := createTestAPIKey(t, db)

	if err := db.DeleteAPIKey(apiKey.ID); err != nil {
		t.Fatalf("Failed to delete API key: %v", err)
	}

	_, err := db.ReadAPIKeyByID(apiKey.ID)
	if err == nil {
		t.Fatal("Expected error when reading deleted API key")
	}
}
