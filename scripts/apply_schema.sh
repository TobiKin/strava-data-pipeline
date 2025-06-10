#!/bin/bash
# Script to apply database schema to an external PostgreSQL server without pg_isready dependency

# Set default values that will be overridden by environment variables
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-""}
DB_NAME=${DB_NAME:-"strava_pipeline"}
DB_SSL_MODE=${DB_SSL_MODE:-"disable"}

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Applying Database Schema ===${NC}"

# Build the DB checker utility if it doesn't exist
if [ ! -f ./bin/dbcheck ] || [ ./cmd/dbcheck/main.go -nt ./bin/dbcheck ]; then
  echo -e "${YELLOW}Building DB connection checker utility...${NC}"
  go build -o ./bin/dbcheck ./cmd/dbcheck >/dev/null 2>&1
fi

# Check database connectivity
if [ -f ./bin/dbcheck ]; then
  echo -e "${YELLOW}Checking database connectivity...${NC}"
  ./bin/dbcheck
  if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Cannot connect to PostgreSQL server at ${DB_HOST}:${DB_PORT}${NC}"
    echo "Please make sure the server is running and accessible."
    exit 1
  else
    echo -e "${GREEN}✅ Connected to PostgreSQL successfully!${NC}"
  fi
else
  echo -e "${YELLOW}⚠️ DB checker utility not available, skipping connection check.${NC}"
  echo -e "${YELLOW}Will attempt to apply schema anyway.${NC}"
fi

# Function to apply schema
apply_schema() {
  local schema_file=$1
  echo -e "${YELLOW}Applying schema from ${schema_file}...${NC}"

  if [[ -z "${DB_PASSWORD}" ]]; then
    # No password
    PGPASSWORD="" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -f "${schema_file}" 2>/tmp/schema_error
  else
    # With password
    PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -f "${schema_file}" 2>/tmp/schema_error
  fi

  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Schema applied successfully!${NC}"
    return 0
  else
    echo -e "${RED}❌ Error applying schema:${NC}"
    cat /tmp/schema_error
    return 1
  fi
}

# Check if psql is available
if command -v psql &>/dev/null; then
  # First check if the database exists and create it if it doesn't
  echo -e "${YELLOW}Checking if database '${DB_NAME}' exists...${NC}"

  if [[ -z "${DB_PASSWORD}" ]]; then
    DB_EXISTS=$(PGPASSWORD="" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -lqt | cut -d \| -f 1 | grep -qw "${DB_NAME}"; echo $?)
  else
    DB_EXISTS=$(PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -lqt | cut -d \| -f 1 | grep -qw "${DB_NAME}"; echo $?)
  fi

  if [ $DB_EXISTS -ne 0 ]; then
    echo -e "${YELLOW}⚠️ Database '${DB_NAME}' does not exist. Creating it now...${NC}"
    if [[ -z "${DB_PASSWORD}" ]]; then
      PGPASSWORD="" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -c "CREATE DATABASE ${DB_NAME};"
    else
      PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -c "CREATE DATABASE ${DB_NAME};"
    fi

    if [ $? -ne 0 ]; then
      echo -e "${RED}❌ Failed to create database.${NC}"
      exit 1
    else
      echo -e "${GREEN}✅ Database created successfully.${NC}"
    fi
  else
    echo -e "${GREEN}✅ Database '${DB_NAME}' already exists.${NC}"
  fi

  # Apply the schema
  apply_schema "./scripts/schema.sql"

  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Schema setup complete!${NC}"
  else
    echo -e "${RED}❌ Schema setup failed.${NC}"
    exit 1
  fi

else
  # psql is not available, use our Go application to handle schema initialization
  echo -e "${YELLOW}⚠️ psql command not found. Using application's built-in schema initialization...${NC}"

  # We'll rely on the application's InitSchema functionality when it starts up
  echo -e "${GREEN}✅ The application will initialize the database schema on startup.${NC}"

  # Create a small Go program to create the database if it doesn't exist
  cat > /tmp/create_db.go <<EOT
package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	// Connect to the postgres database to create our target database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		host, port, user, password, sslmode)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Printf("Error connecting to postgres: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create the database if it doesn't exist
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		// Check if error is because database already exists
		if err.Error() == fmt.Sprintf("pq: database \"%s\" already exists", dbname) {
			fmt.Printf("Database %s already exists, continuing...\n", dbname)
			return
		}
		fmt.Printf("Error creating database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Database %s created successfully\n", dbname)
}
EOT

  echo -e "${YELLOW}Creating database if needed...${NC}"
  go run /tmp/create_db.go

  # Cleanup
  rm /tmp/create_db.go
fi
