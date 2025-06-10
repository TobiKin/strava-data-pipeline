#!/bin/bash
# Start script for the Strava Data Pipeline service

# Set environment variables
export STRAVA_CLIENT_ID="identifier_change_this"
export STRAVA_CLIENT_SECRET="your_strava_client_secret_change_this"
export STRAVA_CALLBACK_URL="http://localhost:8080/auth/callback"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_USER="user"
export DB_PASSWORD="password"
export DB_NAME="tempdb"
export DB_SSL_MODE="disable"
export JWT_SECRET="your_jwt_secret_key_change_this"
export SERVER_PORT="8080"

# Apply database schema using our helper script
echo "🔍 Setting up database..."
./scripts/apply_schema.sh
if [ $? -ne 0 ]; then
  echo "❌ Failed to set up database."

  # Check if Docker is installed and offer to start PostgreSQL in Docker
  if command -v docker &>/dev/null; then
    echo "🐳 Would you like to start PostgreSQL in Docker? (y/n)"
    read -r answer
    if [[ "$answer" =~ ^[Yy] ]]; then
      echo "🐳 Starting PostgreSQL in Docker..."
      docker run --name strava-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB="$DB_NAME" -p 5432:5432 -d postgres

      echo "🕒 Waiting for PostgreSQL to start..."
      sleep 5

      # Try applying schema again
      echo "🔄 Trying to set up database again..."
      ./scripts/apply_schema.sh
      if [ $? -ne 0 ]; then
        echo "❌ Failed to set up database in Docker."
        echo "Please check Docker logs and try again."
        exit 1
      fi

      echo "✅ PostgreSQL is running in Docker and database is set up."
    else
      echo "Please start PostgreSQL manually and try again."
      exit 1
    fi
  else
    echo "Docker not found. Please start PostgreSQL manually and try again."
    exit 1
  fi
fi

# Build the application
echo "🔨 Building the application..."
go build -o ./bin/strava-pipeline ./cmd/server

# Check if build was successful
if [ $? -ne 0 ]; then
  echo "❌ Build failed."
  exit 1
fi

# Run the application
echo "🚀 Starting Strava Data Pipeline..."
./bin/strava-pipeline
