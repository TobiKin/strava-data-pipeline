package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/TobiKin/strava-data-pipeline/internal/auth"
	"github.com/TobiKin/strava-data-pipeline/internal/db"
	"github.com/TobiKin/strava-data-pipeline/internal/strava"
	"github.com/gorilla/mux"
)

// Server represents the API server
type Server struct {
	db           *db.DB
	stravaClient *strava.Client
	authService  *auth.Service
	router       *mux.Router
	templates    *template.Template
}

// New creates a new API server
func New(db *db.DB, stravaClient *strava.Client, authService *auth.Service) *Server {
	s := &Server{
		db:           db,
		stravaClient: stravaClient,
		authService:  authService,
		router:       mux.NewRouter(),
	}

	// Initialize templates
	s.templates = template.Must(template.New("").Parse(templateString))

	s.routes()
	return s
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// routes sets up the routes for the API server
func (s *Server) routes() {
	// Static files
	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Web UI routes
	s.router.HandleFunc("/", s.homeHandler).Methods("GET")
	s.router.HandleFunc("/login", s.loginHandler).Methods("GET")
	s.router.HandleFunc("/dashboard", s.dashboardHandler).Methods("GET")

	// Public routes
	s.router.HandleFunc("/api/health", s.healthHandler).Methods("GET")
	s.router.HandleFunc("/api/auth/strava", s.stravaAuthHandler).Methods("GET")
	s.router.HandleFunc("/api/auth/callback", s.stravaCallbackHandler).Methods("GET")

	// API routes (protected)
	api := s.router.PathPrefix("/api/v1").Subrouter()
	api.Use(s.authService.AuthMiddleware)

	api.HandleFunc("/activities", s.listActivitiesHandler).Methods("GET")
	api.HandleFunc("/activities/{id}", s.getActivityHandler).Methods("GET")

	// Admin routes
	admin := s.router.PathPrefix("/admin").Subrouter()
	admin.Use(s.authService.JWTMiddleware)

	admin.HandleFunc("/keys", s.listKeysHandler).Methods("GET")
	admin.HandleFunc("/keys", s.createKeyHandler).Methods("POST")
	admin.HandleFunc("/sync", s.syncActivitiesHandler).Methods("POST")

	// Serve static files if needed
	// s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}

// Web UI handlers

// homeHandler handles the home page
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title":       "Strava Data Pipeline",
		"CurrentYear": time.Now().Year(),
	}
	s.renderTemplate(w, "home", data)
}

// loginHandler handles the login page
func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Generate the Strava authorization URL
	authURL := s.stravaClient.StartAuthFlow()

	data := map[string]interface{}{
		"Title":       "Login with Strava",
		"AuthURL":     authURL,
		"CurrentYear": time.Now().Year(),
	}
	s.renderTemplate(w, "login", data)
}

// dashboardHandler handles the dashboard page
func (s *Server) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Validate token
	claims, err := s.authService.ValidateJWT(token)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Get user information
	user, err := s.stravaClient.GetUserByID(claims.UserID)
	if err != nil || user == nil {
		http.Error(w, "Error getting user information", http.StatusInternalServerError)
		return
	}

	// Get API keys for user (we'll need to implement this)
	apiKeys, err := s.db.GetAPIKeysForUser(claims.UserID)
	if err != nil {
		http.Error(w, "Error getting API keys", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title":       "Dashboard",
		"User":        user,
		"APIKeys":     apiKeys,
		"Token":       token,
		"CurrentYear": time.Now().Year(),
	}
	s.renderTemplate(w, "dashboard", data)
}

// API handlers

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// stravaAuthHandler initiates the Strava OAuth flow
func (s *Server) stravaAuthHandler(w http.ResponseWriter, r *http.Request) {
	authURL := s.stravaClient.StartAuthFlow()

	// Check if the request prefers HTML (browser) or JSON (API)
	if preferHTML(r) {
		http.Redirect(w, r, authURL, http.StatusFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"auth_url": authURL,
		})
	}
}

// stravaCallbackHandler handles the Strava OAuth callback
func (s *Server) stravaCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange the code for a token
	resp, err := s.stravaClient.HandleAuthCallback(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error exchanging code: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate a JWT token for the user
	token, err := s.authService.GenerateJWT(resp.Athlete.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating token: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if the request prefers HTML (browser) or JSON (API)
	if preferHTML(r) {
		http.Redirect(w, r, "/dashboard?token="+token, http.StatusFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	}
}

// listActivitiesHandler handles requests to list activities
func (s *Server) listActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	// Get activities from the database
	activities, err := s.db.GetActivities(limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting activities: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}

// getActivityHandler handles requests to get a specific activity
func (s *Server) getActivityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid activity ID", http.StatusBadRequest)
		return
	}

	activity, err := s.db.GetActivityByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting activity: %v", err), http.StatusInternalServerError)
		return
	}

	if activity == nil {
		http.Error(w, "Activity not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activity)
}

// Admin handlers

// listKeysHandler handles requests to list API keys
func (s *Server) listKeysHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromContext(r)

	// Get API keys for user
	apiKeys, err := s.db.GetAPIKeysForUser(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting API keys: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apiKeys)
}

// createKeyHandler handles requests to create a new API key
func (s *Server) createKeyHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Description string `json:"description"`
		ExpiryDays  int    `json:"expiry_days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, _ := getUserIDFromContext(r)

	// Generate a new API key
	apiKey, err := s.authService.GenerateAPIKey(req.Description, req.ExpiryDays)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating API key: %v", err), http.StatusInternalServerError)
		return
	}

	// Associate the API key with the user
	if err := s.db.AssociateAPIKeyWithUser(apiKey, userID); err != nil {
		http.Error(w, fmt.Sprintf("Error associating API key with user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"key": apiKey,
	})
}

// syncActivitiesHandler handles requests to manually sync activities
func (s *Server) syncActivitiesHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Days int `json:"days"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Days <= 0 {
		req.Days = 1
	}

	// Start a goroutine to sync activities
	go func() {
		syncStartTime := time.Now().Add(-time.Duration(req.Days) * 24 * time.Hour)
		if err := s.stravaClient.FetchActivities(syncStartTime, 100); err != nil {
			log.Printf("Error syncing activities: %v", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "sync started",
	})
}

// Helper functions

// renderTemplate renders a template with the given data
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/html")

	// Add the template name to the data for base template to select correct content
	if data == nil {
		data = make(map[string]interface{})
	}
	data["Template"] = name
	data["CurrentYear"] = time.Now().Year()

	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
	}
}

// getUserIDFromContext gets the user ID from the request context
func getUserIDFromContext(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value("userID").(int64)
	return userID, ok
}

// preferHTML checks if the request prefers HTML over JSON
func preferHTML(r *http.Request) bool {
	// Check Accept header
	accept := r.Header.Get("Accept")
	if accept != "" {
		return accept == "text/html" || accept == "*/*"
	}

	// Check if it's a browser request
	userAgent := r.Header.Get("User-Agent")
	return userAgent != "" && (r.Method == "GET" || r.Header.Get("Content-Type") == "")
}

// HTML templates for the web UI
const templateString = `
{{define "base"}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Strava Data Pipeline</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <style>
        /* Additional inline styles if needed */
    </style>
        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 20px;
            border-bottom: 1px solid #eee;
        }
        h1 {
            color: #fc5200;
        }
        .btn {
            display: inline-block;
            background-color: #fc5200;
            color: white;
            padding: 10px 20px;
            border-radius: 4px;
            text-decoration: none;
            font-weight: 600;
            transition: background-color 0.2s;
        }
        .btn:hover {
            background-color: #e34a00;
        }
        .card {
            background-color: #fff;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 20px;
            margin-bottom: 20px;
        }
        .activity-list {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
        }
        .footer {
            margin-top: 40px;
            text-align: center;
            color: #888;
            font-size: 0.9em;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px 15px;
            text-align: left;
            border-bottom: 1px solid #eee;
        }
        th {
            background-color: #f8f8f8;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <h1>Strava Data Pipeline</h1>
            <nav>
                <a href="/" class="btn">Home</a>
            </nav>
        </header>

        <main>
            {{if eq .Template "home"}}
                {{template "home-content" .}}
            {{else if eq .Template "login"}}
                {{template "login-content" .}}
            {{else if eq .Template "dashboard"}}
                {{template "dashboard-content" .}}
            {{end}}
        </main>

        <footer class="footer">
            <p>&copy; {{.CurrentYear}} Strava Data Pipeline</p>
        </footer>
    </div>
</body>
</html>
{{end}}

{{define "home"}}
{{template "base" .}}
{{end}}

{{define "home-content"}}
<div class="card">
    <h2>Welcome to Strava Data Pipeline</h2>
    <p>This service downloads your Strava activities and provides an API for accessing the data.</p>
    <p>To get started, please login with your Strava account:</p>
    <p style="margin-top: 20px;">
        <a href="/login" class="btn">Login with Strava</a>
    </p>
</div>
{{end}}

{{define "login"}}
{{template "base" .}}
{{end}}

{{define "login-content"}}
<div class="card">
    <h2>Login with Strava</h2>
    <p>Click the button below to authenticate with your Strava account:</p>
    <p style="margin-top: 20px;">
        <a href="{{.AuthURL}}" class="btn">Connect with Strava</a>
    </p>
</div>
{{end}}

{{define "dashboard"}}
{{template "base" .}}
{{end}}

{{define "dashboard-content"}}
<div class="card">
    <h2>Welcome, {{.User.firstname}} {{.User.lastname}}!</h2>
    <p>Your Strava account is successfully connected.</p>

    <h3 style="margin-top: 20px;">Your API Keys</h3>
    {{if .APIKeys}}
    <table>
        <thead>
            <tr>
                <th>Description</th>
                <th>Key</th>
                <th>Created</th>
                <th>Expires</th>
            </tr>
        </thead>
        <tbody>
            {{range .APIKeys}}
            <tr>
                <td>{{.Description}}</td>
                <td><code>{{.Key}}</code></td>
                <td>{{.CreatedAt}}</td>
                <td>{{if .ExpiresAt}}{{.ExpiresAt}}{{else}}Never{{end}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
    {{else}}
    <p>You don't have any API keys yet.</p>
    {{end}}

    <div style="margin-top: 20px;">
        <h3>Create a New API Key</h3>
        <form id="apiKeyForm" style="margin-top: 10px;">
            <div style="margin-bottom: 15px;">
                <label for="description" style="display: block; margin-bottom: 5px;">Description:</label>
                <input type="text" id="description" name="description" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;" required>
            </div>
            <div style="margin-bottom: 15px;">
                <label for="expiryDays" style="display: block; margin-bottom: 5px;">Expiry (days, 0 for never):</label>
                <input type="number" id="expiryDays" name="expiryDays" min="0" value="30" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
            </div>
            <button type="submit" class="btn">Create API Key</button>
        </form>
        <div id="apiKeyResult" style="margin-top: 15px; display: none; padding: 15px; background-color: #f8f8f8; border-radius: 4px;"></div>
    </div>

    <div style="margin-top: 40px;">
        <h3>Sync Activities</h3>
        <p>Sync your recent activities from Strava:</p>
        <form id="syncForm" style="margin-top: 10px;">
            <div style="margin-bottom: 15px;">
                <label for="days" style="display: block; margin-bottom: 5px;">Days to sync:</label>
                <input type="number" id="days" name="days" min="1" value="7" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;">
            </div>
            <button type="submit" class="btn">Sync Activities</button>
        </form>
        <div id="syncResult" style="margin-top: 15px; display: none; padding: 15px; background-color: #f8f8f8; border-radius: 4px;"></div>
    </div>
</div>

<script>
document.getElementById('apiKeyForm').addEventListener('submit', function(e) {
    e.preventDefault();
    const description = document.getElementById('description').value;
    const expiryDays = parseInt(document.getElementById('expiryDays').value);

    fetch('/admin/keys', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer {{.Token}}'
        },
        body: JSON.stringify({
            description: description,
            expiry_days: expiryDays
        })
    })
    .then(response => response.json())
    .then(data => {
        const resultDiv = document.getElementById('apiKeyResult');
        resultDiv.style.display = 'block';
        resultDiv.innerHTML = '<strong>New API Key Created:</strong><br><code>' + data.key + '</code><br><br>Make sure to save this key as it won\'t be shown again!';
    })
    .catch(error => {
        const resultDiv = document.getElementById('apiKeyResult');
        resultDiv.style.display = 'block';
        resultDiv.innerHTML = 'Error creating API key: ' + error.message;
    });
});

document.getElementById('syncForm').addEventListener('submit', function(e) {
    e.preventDefault();
    const days = parseInt(document.getElementById('days').value);

    fetch('/admin/sync', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer {{.Token}}'
        },
        body: JSON.stringify({
            days: days
        })
    })
    .then(response => response.json())
    .then(data => {
        const resultDiv = document.getElementById('syncResult');
        resultDiv.style.display = 'block';
        resultDiv.innerHTML = 'Sync started! This may take a few minutes depending on how many activities need to be synced.';
    })
    .catch(error => {
        const resultDiv = document.getElementById('syncResult');
        resultDiv.style.display = 'block';
        resultDiv.innerHTML = 'Error starting sync: ' + error.message;
    });
});
</script>
{{end}}
`
