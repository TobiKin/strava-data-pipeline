package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Strava struct {
	ClientID     int
	ClientSecret string
	CallbackURL  string
	AccessToken  string
	RefreshToken string
}

type Server struct {
	Port int
	Host string
}

type Auth struct {
	JWTSecret     string
	TokenDuration int // in minutes
}

// Config holds all configuration for the application
type Config struct {
	Database Database
	Strava   Strava
	Server   Server
	Auth     Auth
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // type of the config file

	// Look in the configPath directory or current directory for the config file
	if configPath != "" {
		viper.AddConfigPath(configPath)
	}
	viper.AddConfigPath(".")

	// Read in environment variables that match
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found; using environment variables and defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Set defaults
	setDefaults()

	// Explicitly bind environment variables
	bindEnvironmentVariables()

	// Unmarshal config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}

// setDefaults sets the default configuration values
func setDefaults() {
	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.name", "strava_data")
	viper.SetDefault("database.sslmode", "disable")

	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")

	// Auth defaults
	viper.SetDefault("auth.tokenduration", 60) // 1 hour
}

// bindEnvironmentVariables explicitly binds environment variables to configuration keys
func bindEnvironmentVariables() {
	// Database bindings
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.name", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSL_MODE")

	// Strava bindings
	viper.BindEnv("strava.client_id", "STRAVA_CLIENT_ID")
	viper.BindEnv("strava.client_secret", "STRAVA_CLIENT_SECRET")
	viper.BindEnv("strava.callback_url", "STRAVA_CALLBACK_URL")
	viper.BindEnv("strava.access_token", "STRAVA_ACCESS_TOKEN")
	viper.BindEnv("strava.refresh_token", "STRAVA_REFRESH_TOKEN")

	// Server bindings
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.host", "SERVER_HOST")

	// Auth bindings
	viper.BindEnv("auth.jwt_secret", "JWT_SECRET")
	viper.BindEnv("auth.token_duration", "TOKEN_DURATION")
}
