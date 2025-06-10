#!/bin/bash
# Script to run Strava Data Pipeline with an external PostgreSQL server

# Set your external PostgreSQL server details here
EXTERNAL_DB_HOST="192.168.64.5"  # Your external PostgreSQL server IP
EXTERNAL_DB_PORT="5432"          # Your PostgreSQL port
EXTERNAL_DB_USER="user"      # Your PostgreSQL username
EXTERNAL_DB_PASSWORD="password"          # Your PostgreSQL password (leave empty if not needed)
EXTERNAL_DB_NAME="tempdb"   # Your database name

# Set Strava API credentials (you should replace these with your actual credentials)
STRAVA_CLIENT_ID="your_strava_client_id_change_this"
STRAVA_CLIENT_SECRET="your_strava_client_secret_change_this"
STRAVA_CALLBACK_URL="http://localhost:8080/auth/callback"

# Set JWT secret for authentication
JWT_SECRET="your_jwt_secret_key_change_this"

# Server port
SERVER_PORT="8080"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Starting Strava Data Pipeline with External PostgreSQL ===${NC}"

# Set environment variables for DB checker to use
export DB_HOST="${EXTERNAL_DB_HOST}"
export DB_PORT="${EXTERNAL_DB_PORT}"
export DB_USER="${EXTERNAL_DB_USER}"
export DB_PASSWORD="${EXTERNAL_DB_PASSWORD}"
export DB_NAME="${EXTERNAL_DB_NAME}"
export DB_SSL_MODE="disable"

# Check if the external PostgreSQL server is reachable
echo -e "${YELLOW}Checking connection to external PostgreSQL server...${NC}"

# Build the DB checker utility if it doesn't exist
if [ ! -f ./bin/dbcheck ] || [ ./cmd/dbcheck/main.go -nt ./bin/dbcheck ]; then
  echo -e "${YELLOW}Building DB connection checker utility...${NC}"
  go build -o ./bin/dbcheck ./cmd/dbcheck >/dev/null 2>&1
fi

if [ -f ./bin/dbcheck ]; then
  ./bin/dbcheck
  if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Cannot connect to external PostgreSQL server at ${EXTERNAL_DB_HOST}:${EXTERNAL_DB_PORT}${NC}"
    echo "Please make sure the server is running and accessible."
    exit 1
  else
    echo -e "${GREEN}✅ Connected to external PostgreSQL server successfully!${NC}"
  fi
else
  echo -e "${YELLOW}⚠️ DB checker utility not available, skipping PostgreSQL check.${NC}"
  echo -e "${YELLOW}The application will attempt to connect to the database itself.${NC}"
fi

# Export environment variables for the application
export DB_HOST="${EXTERNAL_DB_HOST}"
export DB_PORT="${EXTERNAL_DB_PORT}"
export DB_USER="${EXTERNAL_DB_USER}"
export DB_PASSWORD="${EXTERNAL_DB_PASSWORD}"
export DB_NAME="${EXTERNAL_DB_NAME}"
export DB_SSL_MODE="disable"  # Change to "require" if your connection requires SSL

# Apply database schema using our helper script
echo -e "${YELLOW}Applying database schema...${NC}"
./scripts/apply_schema.sh
if [ $? -ne 0 ]; then
  echo -e "${RED}❌ Failed to apply database schema.${NC}"
  exit 1
fi
echo -e "${GREEN}✅ Database setup completed!${NC}"

export STRAVA_CLIENT_ID="${STRAVA_CLIENT_ID}"
export STRAVA_CLIENT_SECRET="${STRAVA_CLIENT_SECRET}"
export STRAVA_CALLBACK_URL="${STRAVA_CALLBACK_URL}"

export JWT_SECRET="${JWT_SECRET}"
export SERVER_PORT="${SERVER_PORT}"

# Build the application
echo -e "${YELLOW}Building the application...${NC}"
go build -o ./bin/strava-pipeline ./cmd/server

if [ $? -ne 0 ]; then
  echo -e "${RED}❌ Build failed.${NC}"
  exit 1
fi

# Run the application with the config file
echo -e "${GREEN}🚀 Starting Strava Data Pipeline...${NC}"
echo -e "${YELLOW}The application will be available at http://localhost:${SERVER_PORT}${NC}"
./bin/strava-pipeline --config=./config.yaml
