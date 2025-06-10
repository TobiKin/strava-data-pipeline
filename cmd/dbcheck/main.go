// Package main provides a simple utility to check database connectivity
package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// Get database connection parameters from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSL_MODE")

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	// Try to connect with retries
	var db *sql.DB
	var err error
	maxRetries := 3
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				fmt.Println("✅ Database connection successful")
				db.Close()
				os.Exit(0)
			}
		}

		if i < maxRetries-1 {
			fmt.Printf("⚠️ Connection attempt %d failed: %v\n", i+1, err)
			fmt.Printf("🔄 Retrying in %v...\n", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	fmt.Printf("❌ Failed to connect to database after %d attempts: %v\n", maxRetries, err)
	os.Exit(1)
}
