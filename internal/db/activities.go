package db

import (
	"fmt"
	"time"
)

var activitySchema = `
CREATE TABLE IF NOT EXISTS activities (
	id BIGINT PRIMARY KEY,
	name TEXT,
	description TEXT,
	type TEXT,
	distance FLOAT,
	moving_time INT,
	elapsed_time INT,
	total_elevation_gain FLOAT,
	start_date TIMESTAMP,
	start_date_local TIMESTAMP,
	timezone TEXT,
	start_latlng TEXT,
	end_latlng TEXT,
	achievement_count INT,
	kudos_count INT,
	comment_count INT,
	athlete_count INT,
	photo_count INT,
	map_id TEXT,
	map_polyline TEXT,
	trainer BOOLEAN,
	commute BOOLEAN,
	manual BOOLEAN,
	private BOOLEAN,
	visibility TEXT,
	flagged BOOLEAN,
	workout_type INT,
	average_speed FLOAT,
	max_speed FLOAT,
	has_heartrate BOOLEAN,
	average_heartrate FLOAT,
	max_heartrate FLOAT,
	elev_high FLOAT,
	elev_low FLOAT,
	upload_id BIGINT,
	upload_id_str TEXT,
	external_id TEXT,
	athlete_id BIGINT,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW()
);`

type Activity struct {
	ID                 int64     `db:"id"`
	Name               string    `db:"name"`
	Description        string    `db:"description"`
	Type               string    `db:"type"`
	Distance           float64   `db:"distance"`
	MovingTime         int       `db:"moving_time"`
	ElapsedTime        int       `db:"elapsed_time"`
	TotalElevationGain float64   `db:"total_elevation_gain"`
	StartDate          time.Time `db:"start_date"`
	StartDateLocal     time.Time `db:"start_date_local"`
	Timezone           string    `db:"timezone"`
	StartLatLng        string    `db:"start_latlng"`
	EndLatLng          string    `db:"end_latlng"`
	AchievementCount   int       `db:"achievement_count"`
	KudosCount         int       `db:"kudos_count"`
	CommentCount       int       `db:"comment_count"`
	AthleteCount       int       `db:"athlete_count"`
	PhotoCount         int       `db:"photo_count"`
	MapID              string    `db:"map_id"`
	MapPolyline        string    `db:"map_polyline"`
	Trainer            bool      `db:"trainer"`
	Commute            bool      `db:"commute"`
	Manual             bool      `db:"manual"`
	Private            bool      `db:"private"`
	Visibility         string    `db:"visibility"`
	Flagged            bool      `db:"flagged"`
	WorkoutType        int       `db:"workout_type"`
	AverageSpeed       float64   `db:"average_speed"`
	MaxSpeed           float64   `db:"max_speed"`
	HasHeartRate       bool      `db:"has_heartrate"`
	AverageHeartRate   float64   `db:"average_heartrate"`
	MaxHeartRate       float64   `db:"max_heartrate"`
	ElevHigh           float64   `db:"elev_high"`
	ElevLow            float64   `db:"elev_low"`
	UploadID           int64     `db:"upload_id"`
	UploadIDStr        string    `db:"upload_id_str"`
	ExternalID         string    `db:"external_id"`
	AthleteID          int64     `db:"athlete_id"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

func (db *DB) CreateActivitySchema() {
	db.MustExec(activitySchema)
}

func (db *DB) CreateActivity(activity Activity) (Activity, error) {
	query := `
		INSERT INTO activities (
			id, name, description, type, distance, moving_time, elapsed_time,
			total_elevation_gain, start_date, start_date_local, timezone,
			start_latlng, end_latlng, achievement_count, kudos_count,
			comment_count, athlete_count, photo_count, map_id, map_polyline,
			trainer, commute, manual, private, visibility, flagged,
			workout_type, average_speed, max_speed,
			has_heartrate, average_heartrate, max_heartrate,
			elev_high, elev_low, upload_id, upload_id_str,
			external_id, athlete_id
		) VALUES (
			:ID, :Name, :Description, :Type, :Distance,
			:MovingTime, :ElapsedTime, :TotalElevationGain,
			:StartDate, :StartDateLocal, :Timezone,
			:StartLatLng, :EndLatLng,
			:AchievementCount, :KudosCount,
			:CommentCount, :AthleteCount,
			:PhotoCount, :MapID,
			:MapPolyline,
			:Trainer, :Commute,
			:Manual, :Private,
			:Visibility,
			:Flagged,
			:WorkoutType,
			:AverageSpeed,
			:MaxSpeed,
			:HasHeartRate,
			:AverageHeartRate,
			:MaxHeartRate,
			:ElevHigh,
			:ElevLow,
			:UploadID,
			:UploadIDStr,
			:ExternalID,
			:AthleteID
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			type = EXCLUDED.type,
			distance = EXCLUDED.distance,
			moving_time = EXCLUDED.moving_time,
			elapsed_time = EXCLUDED.elapsed_time,
			total_elevation_gain = EXCLUDED.total_elevation_gain,
			start_date = EXCLUDED.start_date,
			start_date_local = EXCLUDED.start_date_local,
			timezone = EXCLUDED.timezone,
			start_latlng = EXCLUDED.start_latlng,
			end_latlng = EXCLUDED.end_latlng,
			achievement_count = EXCLUDED.achievement_count,
			kudos_count = EXCLUDED.kudos_count,
			comment_count = EXCLUDED.comment_count,
			athlete_count = EXCLUDED.athlete_count,
			photo_count = EXCLUDED.photo_count,
			map_id = EXCLUDED.map_id,
			map_polyline = EXCLUDED.map_polyline,
			trainer = EXCLUDED.trainer,
			commute = EXCLUDED.commute,
			manual = EXCLUDED.manual,
			private = EXCLUDED.private,
			visibility = EXCLUDED.visibility,
			flagged = EXCLUDED.flagged,
			workout_type = EXCLUDED.workout_type,
			average_speed = EXCLUDED.average_speed,
			max_speed = EXCLUDED.max_speed,
			has_heartrate = EXCLUDED.has_heartrate,
			average_heartrate = EXCLUDED.average_heartrate,
			max_heartrate = EXCLUDED.max_heartrate,
			elev_high = EXCLUDED.elev_high,
			elev_low = EXCLUDED.elev_low,
			upload_id = EXCLUDED.upload_id,
			upload_id_str = EXCLUDED.upload_id_str,
			external_id = EXCLUDED.external_id,
			athlete_id = EXCLUDED.athlete_id,
			updated_at = NOW()
		RETURNING *
	`

	err := db.Get(&activity, query, activity)
	if err != nil {
		return Activity{}, fmt.Errorf("error creating activity: %w", err)
	}

	return activity, nil
}

func (db *DB) GetActivityByID(id int64) (Activity, error) {
	var activity Activity
	query := `
		SELECT * FROM activities WHERE id = $1
	`
	err := db.Get(&activity, query, id)
	if err != nil {
		if isNoRows(err) {
			return Activity{}, fmt.Errorf("no activity found with id %d", id)
		}
		return Activity{}, fmt.Errorf("error retrieving activity: %w", err)
	}
	return activity, nil
}

func (db *DB) GetLastActivities(limit int) ([]Activity, error) {
	var activities []Activity
	query := `
		SELECT * FROM activities
		ORDER BY start_date DESC
		LIMIT $1
	`
	err := db.Select(&activities, query, limit)
	if err != nil {
		return nil, fmt.Errorf("error retrieving last activities: %w", err)
	}
	return activities, nil
}

func (db *DB) UpdateActivity(activity Activity) (Activity, error) {
	query := `
		UPDATE activities
		SET name = :Name, description = :Description, type = :Type,
			distance = :Distance, moving_time = :MovingTime,
			elapsed_time = :ElapsedTime, total_elevation_gain = :TotalElevationGain,
			start_date = :StartDate, start_date_local = :StartDateLocal,
			timezone = :Timezone, start_latlng = :StartLatLng,
			end_latlng = :EndLatLng, achievement_count = :AchievementCount,
			kudos_count = :KudosCount, comment_count = :CommentCount,
			athlete_count = :AthleteCount, photo_count = :PhotoCount,
			map_id = :MapID, map_polyline = :MapPolyline,
			trainer = :Trainer, commute = :Commute, manual = :Manual,
			private = :Private, visibility = :Visibility,
			flagged = :Flagged, workout_type = :WorkoutType,
			average_speed = :AverageSpeed, max_speed = :MaxSpeed,
			has_heartrate = :HasHeartRate, average_heartrate = :AverageHeartRate,
			max_heartrate = :MaxHeartRate, elev_high = :ElevHigh,
			elev_low = :ElevLow, upload_id = :UploadID,
			upload_id_str = :UploadIDStr, external_id = :ExternalID,
			athlete_id = :AthleteID, updated_at = NOW()
		WHERE id = $1
		RETURNING *
	`
	err := db.Get(&activity, query, activity.ID)
	if err != nil {
		return Activity{}, fmt.Errorf("error updating activity: %w", err)
	}
	return activity, nil
}

func (db *DB) DeleteActivity(id int64) error {
	query := `
		DELETE FROM activities WHERE id = $1
	`
	_, err := db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting activity with id %d: %w", id, err)
	}
	return nil
}
