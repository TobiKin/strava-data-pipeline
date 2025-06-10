# Strava Data Pipeline

A Golang service that downloads Strava activities into a SQL database and provides an API for authenticated access.

## Features

- OAuth2 authentication with Strava API
- Automatic syncing of activities
- REST API with authentication for data access
- Dockerized deployment

## Prerequisites

- Go 1.21+
- PostgreSQL (or Docker to run PostgreSQL in a container)
- Strava API credentials (Client ID and Client Secret)

## Setup

### Local Development

1. Clone the repository:
   ```
   git clone https://github.com/TobiKin/strava-data-pipeline.git
   cd strava-data-pipeline
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. Configure the application:
   Edit `config.yaml` and update the following:
   - Strava Client ID and Client Secret
   - Database connection details (if not using defaults)

4. Run the application:

   a. With local PostgreSQL:
   ```
   ./scripts/start.sh
   ```

   b. With external PostgreSQL server:
   ```
   ./scripts/run_external.sh
   ```
   Edit this script first to set your external PostgreSQL server details.

   c. For development with hot reloading (requires [air](https://github.com/cosmtrek/air)):
   ```
   ./scripts/run_dev.sh
   ```

### Docker Deployment

1. Create a `.env` file with your Strava API credentials:
   ```
   STRAVA_CLIENT_ID=your_client_id
   STRAVA_CLIENT_SECRET=your_client_secret
   JWT_SECRET=your_jwt_secret
   ```

2. Start the containers:
   ```
   docker-compose up -d
   ```

## API Endpoints

### Authentication

- `GET /api/auth/strava`: Start the Strava OAuth flow
- `GET /api/auth/callback`: Strava OAuth callback

### Activities

- `GET /api/v1/activities`: List activities
  - Query parameters:
    - `limit`: Number of activities to return (default: 20)
    - `offset`: Pagination offset (default: 0)
  - Required header: `X-API-Key: your_api_key`

- `GET /api/v1/activities/{id}`: Get a specific activity
  - Required header: `X-API-Key: your_api_key`

### Admin

- `GET /admin/keys`: List API keys
  - Required header: `Authorization: Bearer your_jwt_token`

- `POST /admin/keys`: Create a new API key
  - Required header: `Authorization: Bearer your_jwt_token`
  - Request body:
    ```json
    {
      "description": "API key description",
      "expiry_days": 30
    }
    ```

- `POST /admin/sync`: Manually trigger activity sync
  - Required header: `Authorization: Bearer your_jwt_token`
  - Request body:
    ```json
    {
      "days": 7  // Number of days to sync (default: 1)
    }
    ```

## Database Schema

The application uses the following tables:

- `activities`: Stores activity data from Strava
- `users`: Stores user information and OAuth tokens
- `api_keys`: Stores API keys for authentication

## Testing

### Application Tests

You can test the application using the test script:

```
./scripts/test.sh
```

This will:
1. Check if the required tools are installed
2. Verify PostgreSQL is running
3. Test API endpoints to make sure they're working correctly

### Database Tests

To verify your database connection and schema setup:

```
./scripts/test_db.sh
```

This will:
1. Test the PostgreSQL connection
2. Verify that all required tables exist
3. Perform a basic data operation test

## Troubleshooting

### Common Issues

- **Database Connection Errors**: Make sure PostgreSQL is running and the database exists
  - Run `./scripts/test_db.sh` to diagnose database issues
  - Use `./scripts/apply_schema.sh` to ensure the schema is properly set up
- **Authentication Errors**: Verify your Strava API credentials are correct
- **API Key Issues**: Check that the API key is valid and not expired

### Logs

Check the application logs for more detailed error messages. When running in Docker, you can view logs with:

```
docker-compose logs -f app
```

## License

MIT
