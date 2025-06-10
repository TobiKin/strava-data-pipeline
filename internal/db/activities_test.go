package db

import (
	"testing"
	"time"
)

func setupTestActivityDB(t *testing.T) *DB {
	db := setupTestDB(t)
	db.CreateActivitySchema()
	return db
}

func createTestActivity(t *testing.T, db *DB) Activity {
	activity := Activity{
		ID:                 1,
		Name:               "Test Activity",
		Description:        "Test Description",
		Type:               "Run",
		Distance:           10.0,
		MovingTime:         3600,
		ElapsedTime:        3700,
		TotalElevationGain: 100.0,
		StartDate:          time.Now(),
		StartDateLocal:     time.Now(),
		Timezone:           "Europe/Berlin",
		StartLatLng:        "52.52,13.405",
		EndLatLng:          "52.52,13.405",
		AchievementCount:   1,
		KudosCount:         2,
		CommentCount:       0,
		AthleteCount:       1,
		PhotoCount:         0,
		MapID:              "mapid",
		MapPolyline:        "polyline",
		Trainer:            false,
		Commute:            false,
		Manual:             false,
		Private:            false,
		Visibility:         "everyone",
		Flagged:            false,
		WorkoutType:        1,
		AverageSpeed:       2.5,
		MaxSpeed:           3.0,
		HasHeartRate:       true,
		AverageHeartRate:   150.0,
		MaxHeartRate:       170.0,
		ElevHigh:           50.0,
		ElevLow:            10.0,
		UploadID:           12345,
		UploadIDStr:        "12345",
		ExternalID:         "extid",
		AthleteID:          1,
	}
	created, err := db.CreateActivity(activity)
	if err != nil {
		t.Fatalf("Failed to create test activity: %v", err)
	}
	return created
}

func TestCreateActivity(t *testing.T) {
	db := setupTestActivityDB(t)
	defer db.Close()
	_ = createTestActivity(t, db)
}

func TestGetActivityByID(t *testing.T) {
	db := setupTestActivityDB(t)
	defer db.Close()
	created := createTestActivity(t, db)
	activity, err := db.GetActivityByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get activity by ID: %v", err)
	}
	if activity.ID != created.ID {
		t.Fatalf("Expected activity ID %d, got %d", created.ID, activity.ID)
	}
}

func TestUpdateActivity(t *testing.T) {
	db := setupTestActivityDB(t)
	defer db.Close()
	created := createTestActivity(t, db)
	created.Name = "Updated Name"
	updated, err := db.UpdateActivity(created)
	if err != nil {
		t.Fatalf("Failed to update activity: %v", err)
	}
	if updated.Name != "Updated Name" {
		t.Fatalf("Expected updated name, got %s", updated.Name)
	}
}

func TestDeleteActivity(t *testing.T) {
	db := setupTestActivityDB(t)
	defer db.Close()
	created := createTestActivity(t, db)
	err := db.DeleteActivity(created.ID)
	if err != nil {
		t.Fatalf("Failed to delete activity: %v", err)
	}
	_, err = db.GetActivityByID(created.ID)
	if err == nil {
		t.Fatal("Expected error for deleted activity, got nil")
	}
}
