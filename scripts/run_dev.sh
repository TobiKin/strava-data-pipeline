#!/bin/bash

# This script runs the application in development mode

# Your external PostgreSQL server details
DB_HOST="192.168.64.5"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD=""  # Leave empty if not needed
DB_NAME="strava_data"

# Set environment variables for DB connection
export DB_HOST="${DB_HOST}"
export DB_PORT="${DB_PORT}"
export DB_USER="${DB_USER}"
export DB_PASSWORD="${DB_PASSWORD}"
export DB_NAME="${DB_NAME}"
export DB_SSL_MODE="disable"

# Build the DB checker utility if it doesn't exist
if [ ! -f ./bin/dbcheck ] || [ ./cmd/dbcheck/main.go -nt ./bin/dbcheck ]; then
    echo "Building DB connection checker utility..."
    go build -o ./bin/dbcheck ./cmd/dbcheck >/dev/null 2>&1
fi

# Ensure the database is running
echo "Checking if external PostgreSQL is running at ${DB_HOST}:${DB_PORT}..."
if [ -f ./bin/dbcheck ]; then
    ./bin/dbcheck
    if [ $? -ne 0 ]; then
        echo "Cannot connect to PostgreSQL server at ${DB_HOST}:${DB_PORT}"
        echo "Please make sure your external PostgreSQL server is running and accessible"
        exit 1
    else
        echo "Successfully connected to PostgreSQL at ${DB_HOST}:${DB_PORT}"
    fi
else
    echo "Warning: DB checker utility not available. Skipping connection check."
fi

# Set environment variables
export DB_HOST="${DB_HOST}"
export DB_PORT="${DB_PORT}"
export DB_USER="${DB_USER}"
export DB_PASSWORD="${DB_PASSWORD}"
export DB_NAME="${DB_NAME}"
export DB_SSL_MODE="disable"

# Apply schema using our helper script
echo "Applying database schema..."
./scripts/apply_schema.sh

# Run the application in development mode
echo "Starting the application in development mode..."
go run ./cmd/server/main.go --config=./config.yaml
