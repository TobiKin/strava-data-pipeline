package strava

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/TobiKin/strava-data-pipeline/internal/config"
	"github.com/TobiKin/strava-data-pipeline/internal/db"
	strava "github.com/strava/go.strava"
)

// Client is a wrapper around the Strava API client
type Client struct {
	config        *config.Config
	client        *strava.Client
	authenticator strava.OAuthAuthenticator
	db            *db.DB
}

// New creates a new Strava client
func New(config *config.Config, database *db.DB) (*Client, error) {
	// Create a new authenticator
	authenticator := strava.OAuthAuthenticator{
		CallbackURL: config.Strava.CallbackURL,
	}

	// Set the client ID and secret
	strava.ClientId = config.Strava.ClientID
	strava.ClientSecret = config.Strava.ClientSecret

	// Create a new client with the saved access token
	client := strava.NewClient(config.Strava.AccessToken)

	return &Client{
		config:        config,
		client:        client,
		authenticator: authenticator,
		db:            database,
	}, nil
}

// FetchActivities fetches activities from Strava and stores them in the database
func (c *Client) FetchActivities(after time.Time, limit int) error {
	// Convert time to int64
	afterUnix := after.Unix()

	// Get activities from Strava
	service := strava.NewCurrentAthleteService(c.client)
	activities, err := service.ListActivities().
		After(int(afterUnix)).
		Page(1).
		PerPage(limit).
		Do()

	if err != nil {
		return fmt.Errorf("error fetching activities: %w", err)
	}

	log.Printf("Fetched %d activities from Strava", len(activities))

	// Save activities to the database
	for _, activity := range activities {
		// Convert the activity to a map
		activityMap, err := activityToMap(activity)
		if err != nil {
			log.Printf("Error converting activity to map: %v", err)
			continue
		}

		// Save the activity to the database
		if err := c.db.SaveActivity(activityMap); err != nil {
			log.Printf("Error saving activity: %v", err)
			continue
		}
	}

	return nil
}

// activityToMap converts a Strava activity to a map
func activityToMap(activity *strava.ActivitySummary) (map[string]interface{}, error) {
	// Convert the activity to JSON
	data, err := json.Marshal(activity)
	if err != nil {
		return nil, fmt.Errorf("error marshaling activity: %w", err)
	}

	// Convert JSON to a map
	var activityMap map[string]interface{}
	if err := json.Unmarshal(data, &activityMap); err != nil {
		return nil, fmt.Errorf("error unmarshaling activity: %w", err)
	}

	return activityMap, nil
}

// RefreshToken refreshes the Strava API tokens
func (c *Client) RefreshToken(refreshToken string) (*strava.AuthorizationResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Use the OAuth service to refresh the token
	resp, err := c.authenticator.Authorize(refreshToken, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}

	// Update the client with the new access token
	c.client = strava.NewClient(resp.AccessToken)

	// TODO: Save the new tokens to the configuration or database
	c.config.Strava.AccessToken = resp.AccessToken

	log.Println("Strava API token refreshed")
	return resp, nil
}

// StartAuthFlow starts the OAuth2 authentication flow
func (c *Client) StartAuthFlow() string {
	// The scope determines what the app can access
	// We use the "view_private" permission to access all activities
	authURL := c.authenticator.AuthorizationURL("strava_state", strava.Permissions.ViewPrivate, false)
	return authURL
}

// HandleAuthCallback handles the OAuth2 callback
func (c *Client) HandleAuthCallback(ctx context.Context, code string) (*strava.AuthorizationResponse, error) {
	// Exchange authorization code for token
	resp, err := c.authenticator.Authorize(code, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("error exchanging code for token: %w", err)
	}

	// Update the client with the new access token
	c.client = strava.NewClient(resp.AccessToken)

	// Save the tokens to the config
	c.config.Strava.AccessToken = resp.AccessToken

	// Save user information to the database
	err = c.saveAthlete(&resp.Athlete, resp.AccessToken, "", 0) // No refresh token in the API
	if err != nil {
		log.Printf("Error saving athlete: %v", err)
	}

	return resp, nil
}

// saveAthlete saves athlete information to the database
func (c *Client) saveAthlete(athlete *strava.AthleteDetailed, accessToken, refreshToken string, expiresAt int64) error {
	query := `
		INSERT INTO users (
			id, username, firstname, lastname, city, country, sex,
			access_token, refresh_token, token_expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, to_timestamp($10)
		) ON CONFLICT (id) DO UPDATE SET
			firstname = $3,
			lastname = $4,
			city = $5,
			country = $6,
			sex = $7,
			access_token = $8,
			refresh_token = $9,
			token_expires_at = to_timestamp($10),
			updated_at = NOW()
	`

	_, err := c.db.Exec(
		query,
		athlete.Id,
		"", // No username in the API
		athlete.FirstName,
		athlete.LastName,
		athlete.City,
		athlete.Country,
		string(athlete.Gender),
		accessToken,
		refreshToken,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("error saving athlete: %w", err)
	}

	return nil
}

// GetUserByID retrieves user information from the database
func (c *Client) GetUserByID(userID int64) (map[string]interface{}, error) {
	query := `
		SELECT id, username, firstname, lastname, city, country, sex,
			created_at, updated_at, token_expires_at
		FROM users
		WHERE id = $1
	`

	rows, err := c.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying user: %w", err)
	}
	defer rows.Close()

	results, err := rowsToMaps(rows)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil // Not found
	}

	return results[0], nil
}

// rowsToMaps converts SQL rows to a slice of maps
func rowsToMaps(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		pointers := make([]interface{}, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}

		// Scan the result into the pointers
		if err := rows.Scan(pointers...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			rowMap[colName] = values[i]
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// StartSyncJob starts a job to sync activities from Strava
func (c *Client) StartSyncJob(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			<-ticker.C
			// Sync activities from the last 24 hours
			err := c.FetchActivities(time.Now().Add(-24*time.Hour), 100)
			if err != nil {
				log.Printf("Error syncing activities: %v", err)
			}
		}
	}()
}
