package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TobiKin/strava-data-pipeline/internal/config"
	"github.com/TobiKin/strava-data-pipeline/internal/db"
	"github.com/dgrijalva/jwt-go"
)

// Service provides authentication functionality
type Service struct {
	config *config.Config
	db     *db.DB
}

// New creates a new authentication service
func New(config *config.Config, database *db.DB) *Service {
	return &Service{
		config: config,
		db:     database,
	}
}

// Claims represents the JWT claims
type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.StandardClaims
}

// GenerateAPIKey generates a new API key
func (s *Service) GenerateAPIKey(description string, expiryDays int) (string, error) {
	// Generate a random key
	key, err := generateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("error generating API key: %w", err)
	}

	// Set expiry date if specified
	var expiresAt *string
	if expiryDays > 0 {
		exp := time.Now().AddDate(0, 0, expiryDays).Format(time.RFC3339)
		expiresAt = &exp
	}

	// Save API key to database
	err = s.db.CreateAPIKey(key, description, expiresAt)
	if err != nil {
		return "", fmt.Errorf("error saving API key: %w", err)
	}

	return key, nil
}

// ValidateAPIKey validates an API key
func (s *Service) ValidateAPIKey(key string) (bool, error) {
	return s.db.ValidateAPIKey(key)
}

// generateRandomString generates a random string of the given length
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

// AuthMiddleware is a middleware function that validates API keys
func (s *Service) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header or query parameter
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = r.URL.Query().Get("api_key")
		}

		if apiKey == "" {
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		// Validate API key
		valid, err := s.ValidateAPIKey(apiKey)
		if err != nil {
			http.Error(w, "Error validating API key", http.StatusInternalServerError)
			return
		}

		if !valid {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// API key is valid, call next handler
		next.ServeHTTP(w, r)
	})
}

// GenerateJWT generates a JWT token for the given user ID
func (s *Service) GenerateJWT(userID int64) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.config.Auth.TokenDuration) * time.Minute)

	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Auth.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("error generating JWT token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token
func (s *Service) ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Auth.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// JWTMiddleware is a middleware function that validates JWT tokens
func (s *Service) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get JWT token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if the auth header is in the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate JWT token
		claims, err := s.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		r = r.WithContext(ctx)

		// Token is valid, call next handler
		next.ServeHTTP(w, r)
	})
}
