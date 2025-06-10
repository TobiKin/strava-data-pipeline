package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/TobiKin/strava-data-pipeline/internal/api"
	"github.com/TobiKin/strava-data-pipeline/internal/auth"
	"github.com/TobiKin/strava-data-pipeline/internal/config"
	"github.com/TobiKin/strava-data-pipeline/internal/db"
	"github.com/TobiKin/strava-data-pipeline/internal/strava"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database connection
	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer database.Close()

	// Initialize database schema
	if err := database.InitSchema(); err != nil {
		log.Fatalf("Error initializing database schema: %v", err)
	}

	// Initialize Strava client
	stravaClient, err := strava.New(cfg, database)
	if err != nil {
		log.Fatalf("Error creating Strava client: %v", err)
	}

	// Initialize authentication service
	authService := auth.New(cfg, database)

	// Initialize API server
	apiServer := api.New(database, stravaClient, authService)

	// Start background sync job
	stravaClient.StartSyncJob(1 * time.Hour) // Sync every hour

	// Start HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      apiServer,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting server: %v", err)
	}
}
