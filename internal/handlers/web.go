package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"log"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/service"
	"github.com/guisithos/go-ride-names/internal/strava"
)

type WebHandler struct {
	sessions     *auth.SessionStore
	oauthCfg     *auth.OAuth2Config
	stravaConfig *config.Config
}

func NewWebHandler(sessions *auth.SessionStore, oauthCfg *auth.OAuth2Config, stravaConfig *config.Config) *WebHandler {
	return &WebHandler{
		sessions:     sessions,
		oauthCfg:     oauthCfg,
		stravaConfig: stravaConfig,
	}
}

func (h *WebHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handleHome)
	mux.HandleFunc("/dashboard", h.handleDashboard)
	mux.HandleFunc("/rename-activities", h.handleRenameActivities)
	mux.HandleFunc("/subscribe", h.handleSubscribe)
	mux.HandleFunc("/subscription-status", h.handleSubscriptionStatus)
}

func (h *WebHandler) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
		<html>
			<head>
				<title>Go Ride Names - Make Your Strava Activities Fun!</title>
				<style>
					body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
					.container { text-align: center; }
					.btn { display: inline-block; background: #FC4C02; color: white; padding: 12px 24px; 
						   text-decoration: none; border-radius: 4px; margin: 20px 0; }
					.features { text-align: left; margin: 20px 0; }
					.feature-list { list-style-type: none; padding: 0; }
					.feature-list li { margin: 10px 0; padding: 10px; background: #f5f5f5; border-radius: 4px; }
				</style>
			</head>
			<body>
				<div class="container">
					<h1>Go Ride Names</h1>
					<p>Automatically rename your Strava activities with fun, witty names!</p>
					
					<div class="features">
						<h2>What does it do?</h2>
						<ul class="feature-list">
							<li>Finds your activities with default names</li>
							<li>Generates activity-specific witty names</li>
							<li>Updates them automatically</li>
							<li>Works with runs, rides, swims, and more!</li>
						</ul>
					</div>

					<a href="/auth" class="btn">Connect with Strava</a>
				</div>
			</body>
		</html>
	`)
}

func (h *WebHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tokens, exists := h.sessions.GetTokens("user")
	if !exists {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<html>
			<head>
				<title>Dashboard - Go Ride Names</title>
				<style>
					body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
					.activity { border: 1px solid #ddd; padding: 10px; margin: 10px 0; border-radius: 4px; }
					.btn { 
						background: #FC4C02; 
						color: white; 
						padding: 12px 24px; 
						text-decoration: none; 
						border-radius: 4px; 
						border: none; 
						cursor: pointer;
						margin-left: 10px;
					}
					.btn:disabled {
						background: #ccc;
						cursor: not-allowed;
					}
					.loading { text-align: center; padding: 20px; }
					.error { color: red; padding: 10px; }
					.header { 
						display: flex; 
						justify-content: space-between; 
						align-items: center; 
						margin-bottom: 20px; 
					}
					.buttons-container {
						display: flex;
						gap: 10px;
					}
					.status { 
						padding: 10px; 
						margin: 20px 0; 
						border-radius: 4px; 
					}
					.status.active { background: #e8f5e9; color: #2e7d32; }
					.status.inactive { background: #ffebee; color: #c62828; }
				</style>
			</head>
			<body>
				<div class="header">
					<h1>Your Activities</h1>
					<div class="buttons-container">
						<button id="rename" class="btn">Rename All</button>
						<button id="subscribe" class="btn">Activate Auto-Rename</button>
					</div>
				</div>
				<div id="subscription-status" class="status inactive">
					Auto-rename is currently inactive
				</div>
				<div id="activities">
					<div class="loading">Loading activities...</div>
				</div>

				<script>
					// Function to format date
					function formatDate(dateStr) {
						const date = new Date(dateStr);
						return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
					}

					// Function to format distance in km
					function formatDistance(meters) {
						return (meters / 1000).toFixed(2) + ' km';
					}

					// Function to load activities
					async function loadActivities() {
						try {
							const response = await fetch('https://www.strava.com/api/v3/athlete/activities', {
								headers: {
									'Authorization': 'Bearer %s'
								}
							});
							
							if (!response.ok) {
								throw new Error('Failed to fetch activities');
							}

							const activities = await response.json();
							const container = document.getElementById('activities');
							container.innerHTML = ''; // Clear loading message

							activities.forEach(activity => {
								const div = document.createElement('div');
								div.className = 'activity';
								div.innerHTML = `+"`"+`
									<h3>${activity.name}</h3>
									<p>Type: ${activity.type}</p>
									<p>Distance: ${formatDistance(activity.distance)}</p>
									<p>Date: ${formatDate(activity.start_date_local)}</p>
								`+"`"+`;
								container.appendChild(div);
							});
						} catch (error) {
							const container = document.getElementById('activities');
							container.innerHTML = '<div class="error">Error loading activities: ' + error.message + '</div>';
						}
					}

					// Load activities when page loads
					loadActivities();

					// Handle rename button click
					document.getElementById('rename').addEventListener('click', async () => {
						const button = document.getElementById('rename');
						button.disabled = true;
						button.textContent = 'Renaming...';

						try {
							await fetch('/rename-activities', {
								method: 'POST',
								headers: {
									'Authorization': 'Bearer %s'
								}
							});
							
							// Reload activities after renaming
							loadActivities();
						} catch (error) {
							alert('Error renaming activities: ' + error.message);
						} finally {
							button.disabled = false;
							button.textContent = 'Rename Activities';
						}
					});

					// Add subscription handling
					document.getElementById('subscribe').addEventListener('click', async () => {
						const button = document.getElementById('subscribe');
						const status = document.getElementById('subscription-status');
						
						button.disabled = true;
						button.textContent = 'Activating...';

						try {
							const response = await fetch('/subscribe', {
								method: 'POST',
								headers: {
									'Authorization': 'Bearer %s'
								}
							});
							
							if (response.ok) {
								status.className = 'status active';
								status.textContent = 'Auto-rename is active! New activities will be renamed automatically.';
								button.style.display = 'none';
							} else {
								throw new Error('Failed to activate');
							}
						} catch (error) {
							alert('Error activating auto-rename: ' + error.message);
							button.disabled = false;
							button.textContent = 'Activate Auto-Rename';
						}
					});
				</script>
			</body>
		</html>
	`, tokens.AccessToken, tokens.AccessToken, tokens.AccessToken)
}

func (h *WebHandler) handleRenameActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokens, exists := h.sessions.GetTokens("user")
	if !exists {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken, h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
	activityService := service.NewActivityService(client)

	_, err := activityService.ListActivities(1, 30, 0, 0, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating activities: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokens, exists := h.sessions.GetTokens("user")
	if !exists {
		log.Printf("No tokens found in session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get base URL from request or environment
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://" + r.Host
	}
	callbackURL := baseURL + "/webhook"
	log.Printf("Subscription attempt - Base URL: %s, Callback URL: %s", baseURL, callbackURL)

	verifyToken := os.Getenv("WEBHOOK_VERIFY_TOKEN")
	if verifyToken == "" {
		log.Printf("Error: WEBHOOK_VERIFY_TOKEN not configured")
		http.Error(w, "Webhook verify token not configured", http.StatusInternalServerError)
		return
	}

	log.Printf("Creating Strava client with ID: %s", h.stravaConfig.StravaClientID)
	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken, h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
	activityService := service.NewActivityService(client)

	err := activityService.SubscribeToWebhooks(callbackURL, verifyToken)
	if err != nil {
		log.Printf("Error creating webhook subscription: %v", err)
		http.Error(w, fmt.Sprintf("Error creating subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Store subscription status in session
	if err := h.sessions.Set("webhook_active", true); err != nil {
		log.Printf("Error storing webhook status: %v", err)
	}

	log.Printf("Successfully created webhook subscription")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": true})
}

func (h *WebHandler) handleSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	active := h.sessions.Get("webhook_active") != nil
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": active})
}
