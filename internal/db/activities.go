package db

import (
	"fmt"
	"time"
)

// GetActivities retrieves activities from the database
func (db *DB) GetActivities(limit, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT * FROM activities
		ORDER BY start_date DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error querying activities: %w", err)
	}
	defer rows.Close()

	return rowsToMaps(rows)
}

// GetActivityByID retrieves a single activity by ID
func (db *DB) GetActivityByID(id int64) (map[string]interface{}, error) {
	query := `SELECT * FROM activities WHERE id = $1`
	rows, err := db.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("error querying activity: %w", err)
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

// GetActivitiesForUser retrieves activities from the database for a specific user
func (db *DB) GetActivitiesForUser(userID int64, limit, offset int) ([]map[string]interface{}, error) {
	query := `
		SELECT * FROM activities
		WHERE athlete_id = $1
		ORDER BY start_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error querying activities for user: %w", err)
	}
	defer rows.Close()

	return rowsToMaps(rows)
}

// CountActivities counts the total number of activities in the database
func (db *DB) CountActivities() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM activities").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting activities: %w", err)
	}
	return count, nil
}

// CountActivitiesForUser counts the total number of activities for a specific user
func (db *DB) CountActivitiesForUser(userID int64) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM activities WHERE athlete_id = $1", userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting activities for user: %w", err)
	}
	return count, nil
}

// SaveActivity saves an activity to the database
func (db *DB) SaveActivity(activity map[string]interface{}) error {
	// Prepare the query with all the possible fields from the Strava API
	query := `
		INSERT INTO activities (
			id, name, description, type, distance, moving_time, elapsed_time,
			total_elevation_gain, start_date, start_date_local, timezone,
			start_latlng, end_latlng, achievement_count, kudos_count,
			comment_count, athlete_count, photo_count, map_id, map_polyline,
			trainer, commute, manual, private, visibility, flagged, workout_type,
			average_speed, max_speed, has_heartrate, average_heartrate,
			max_heartrate, elev_high, elev_low, upload_id, upload_id_str,
			external_id, athlete_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29,
			$30, $31, $32, $33, $34, $35, $36, $37, $38
		) ON CONFLICT (id) DO UPDATE SET
			name = $2,
			description = $3,
			type = $4,
			distance = $5,
			moving_time = $6,
			elapsed_time = $7,
			total_elevation_gain = $8,
			start_date = $9,
			start_date_local = $10,
			timezone = $11,
			start_latlng = $12,
			end_latlng = $13,
			achievement_count = $14,
			kudos_count = $15,
			comment_count = $16,
			athlete_count = $17,
			photo_count = $18,
			map_id = $19,
			map_polyline = $20,
			trainer = $21,
			commute = $22,
			manual = $23,
			private = $24,
			visibility = $25,
			flagged = $26,
			workout_type = $27,
			average_speed = $28,
			max_speed = $29,
			has_heartrate = $30,
			average_heartrate = $31,
			max_heartrate = $32,
			elev_high = $33,
			elev_low = $34,
			upload_id = $35,
			upload_id_str = $36,
			external_id = $37,
			athlete_id = $38,
			updated_at = NOW()
	`

	// Extract values from the map
	var (
		id                 int64
		name               string
		description        string
		activityType       string
		distance           float64
		movingTime         int
		elapsedTime        int
		totalElevationGain float64
		startDate          time.Time
		startDateLocal     time.Time
		timezone           string
		startLatlng        string
		endLatlng          string
		achievementCount   int
		kudosCount         int
		commentCount       int
		athleteCount       int
		photoCount         int
		mapID              string
		mapPolyline        string
		trainer            bool
		commute            bool
		manual             bool
		private            bool
		visibility         string
		flagged            bool
		workoutType        int
		averageSpeed       float64
		maxSpeed           float64
		hasHeartrate       bool
		averageHeartrate   float64
		maxHeartrate       float64
		elevHigh           float64
		elevLow            float64
		uploadID           int64
		uploadIDStr        string
		externalID         string
		athleteID          int64
	)

	// Extract values from the map with safe type conversions
	if val, ok := activity["id"]; ok {
		if f, ok := val.(float64); ok {
			id = int64(f)
		}
	}

	if val, ok := activity["name"]; ok {
		name, _ = val.(string)
	}

	if val, ok := activity["description"]; ok {
		description, _ = val.(string)
	}

	if val, ok := activity["type"]; ok {
		activityType, _ = val.(string)
	}

	if val, ok := activity["distance"]; ok {
		distance, _ = val.(float64)
	}

	if val, ok := activity["moving_time"]; ok {
		if f, ok := val.(float64); ok {
			movingTime = int(f)
		}
	}

	if val, ok := activity["elapsed_time"]; ok {
		if f, ok := val.(float64); ok {
			elapsedTime = int(f)
		}
	}

	if val, ok := activity["total_elevation_gain"]; ok {
		totalElevationGain, _ = val.(float64)
	}

	// Handle time fields (assuming they're in ISO 8601 format as strings)
	if val, ok := activity["start_date"]; ok {
		if s, ok := val.(string); ok {
			startDate, _ = time.Parse(time.RFC3339, s)
		}
	}

	if val, ok := activity["start_date_local"]; ok {
		if s, ok := val.(string); ok {
			startDateLocal, _ = time.Parse(time.RFC3339, s)
		}
	}

	if val, ok := activity["timezone"]; ok {
		timezone, _ = val.(string)
	}

	// Handle array fields (converting to string representation)
	if val, ok := activity["start_latlng"]; ok {
		if arr, ok := val.([]interface{}); ok && len(arr) >= 2 {
			startLatlng = fmt.Sprintf("[%v,%v]", arr[0], arr[1])
		}
	}

	if val, ok := activity["end_latlng"]; ok {
		if arr, ok := val.([]interface{}); ok && len(arr) >= 2 {
			endLatlng = fmt.Sprintf("[%v,%v]", arr[0], arr[1])
		}
	}

	if val, ok := activity["achievement_count"]; ok {
		if f, ok := val.(float64); ok {
			achievementCount = int(f)
		}
	}

	if val, ok := activity["kudos_count"]; ok {
		if f, ok := val.(float64); ok {
			kudosCount = int(f)
		}
	}

	if val, ok := activity["comment_count"]; ok {
		if f, ok := val.(float64); ok {
			commentCount = int(f)
		}
	}

	if val, ok := activity["athlete_count"]; ok {
		if f, ok := val.(float64); ok {
			athleteCount = int(f)
		}
	}

	if val, ok := activity["photo_count"]; ok {
		if f, ok := val.(float64); ok {
			photoCount = int(f)
		}
	}

	// Handle map object
	if mapObj, ok := activity["map"].(map[string]interface{}); ok {
		if val, ok := mapObj["id"]; ok {
			mapID, _ = val.(string)
		}

		if val, ok := mapObj["summary_polyline"]; ok {
			mapPolyline, _ = val.(string)
		}
	}

	if val, ok := activity["trainer"]; ok {
		trainer, _ = val.(bool)
	}

	if val, ok := activity["commute"]; ok {
		commute, _ = val.(bool)
	}

	if val, ok := activity["manual"]; ok {
		manual, _ = val.(bool)
	}

	if val, ok := activity["private"]; ok {
		private, _ = val.(bool)
	}

	if val, ok := activity["visibility"]; ok {
		visibility, _ = val.(string)
	}

	if val, ok := activity["flagged"]; ok {
		flagged, _ = val.(bool)
	}

	if val, ok := activity["workout_type"]; ok {
		if f, ok := val.(float64); ok {
			workoutType = int(f)
		}
	}

	if val, ok := activity["average_speed"]; ok {
		averageSpeed, _ = val.(float64)
	}

	if val, ok := activity["max_speed"]; ok {
		maxSpeed, _ = val.(float64)
	}

	if val, ok := activity["has_heartrate"]; ok {
		hasHeartrate, _ = val.(bool)
	}

	if val, ok := activity["average_heartrate"]; ok {
		averageHeartrate, _ = val.(float64)
	}

	if val, ok := activity["max_heartrate"]; ok {
		maxHeartrate, _ = val.(float64)
	}

	if val, ok := activity["elev_high"]; ok {
		elevHigh, _ = val.(float64)
	}

	if val, ok := activity["elev_low"]; ok {
		elevLow, _ = val.(float64)
	}

	if val, ok := activity["upload_id"]; ok {
		if f, ok := val.(float64); ok {
			uploadID = int64(f)
		}
	}

	if val, ok := activity["upload_id_str"]; ok {
		uploadIDStr, _ = val.(string)
	}

	if val, ok := activity["external_id"]; ok {
		externalID, _ = val.(string)
	}

	if val, ok := activity["athlete"].(map[string]interface{})["id"]; ok {
		if f, ok := val.(float64); ok {
			athleteID = int64(f)
		}
	}

	// Execute the query
	_, err := db.Exec(
		query,
		id, name, description, activityType, distance, movingTime, elapsedTime,
		totalElevationGain, startDate, startDateLocal, timezone,
		startLatlng, endLatlng, achievementCount, kudosCount,
		commentCount, athleteCount, photoCount, mapID, mapPolyline,
		trainer, commute, manual, private, visibility, flagged, workoutType,
		averageSpeed, maxSpeed, hasHeartrate, averageHeartrate,
		maxHeartrate, elevHigh, elevLow, uploadID, uploadIDStr,
		externalID, athleteID,
	)

	if err != nil {
		return fmt.Errorf("error saving activity: %w", err)
	}

	return nil
}
