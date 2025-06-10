#!/bin/bash
# Test script to verify database connectivity and schema

# Colors
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Strava Data Pipeline Database Test ===${NC}"

# Check if environment variables are set
echo -e "${YELLOW}Checking environment variables...${NC}"
if [[ -z "$DB_HOST" || -z "$DB_PORT" || -z "$DB_USER" || -z "$DB_NAME" ]]; then
  echo -e "${RED}❌ Missing required environment variables.${NC}"
  echo "Please set DB_HOST, DB_PORT, DB_USER, and DB_NAME"
  exit 1
fi

echo -e "${GREEN}✓ Environment variables set${NC}"
echo "  DB_HOST: $DB_HOST"
echo "  DB_PORT: $DB_PORT"
echo "  DB_USER: $DB_USER"
echo "  DB_NAME: $DB_NAME"

# Create a temporary Go test file
cat > /tmp/db_test.go <<EOT
package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Get database connection parameters from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	if sslmode == "" {
		sslmode = "disable"
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Try to connect
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Error opening database connection: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Ping database
	err = db.Ping()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Database connection successful")

	// Check tables
	tables := []string{"activities", "users", "api_keys"}
	for _, table := range tables {
		var exists bool
		query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)"
		err = db.QueryRow(query, table).Scan(&exists)

		if err != nil {
			fmt.Printf("Error checking if table %s exists: %v\n", table, err)
			os.Exit(1)
		}

		if exists {
			fmt.Printf("✓ Table '%s' exists\n", table)
		} else {
			fmt.Printf("❌ Table '%s' does not exist\n", table)
			os.Exit(1)
		}
	}

	// Test inserting a dummy user (for testing only)
	testUser := fmt.Sprintf("test_user_%d", time.Now().Unix())
	_, err = db.Exec("INSERT INTO users (id, username, firstname, lastname) VALUES ($1, $2, $3, $4)",
		time.Now().Unix(), testUser, "Test", "User")

	if err != nil {
		fmt.Printf("Error inserting test user: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Test user inserted successfully")

	// Cleanup test user
	_, err = db.Exec("DELETE FROM users WHERE username = $1", testUser)
	if err != nil {
		fmt.Printf("Warning: Could not clean up test user: %v\n", err)
	} else {
		fmt.Println("✓ Test user cleaned up successfully")
	}

	fmt.Println("✓ All tests passed!")
}
EOT

echo -e "${YELLOW}Running database tests...${NC}"
go run /tmp/db_test.go

# Capture result
TEST_RESULT=$?

# Cleanup
rm /tmp/db_test.go

if [ $TEST_RESULT -eq 0 ]; then
  echo -e "${GREEN}✅ Database connectivity and schema tests passed!${NC}"
else
  echo -e "${RED}❌ Database tests failed.${NC}"
  exit 1
fi
