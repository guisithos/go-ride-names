package handlers

import (
	"fmt"
	"net/http"

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
					.btn { background: #FC4C02; color: white; padding: 12px 24px; 
						   text-decoration: none; border-radius: 4px; border: none; cursor: pointer; }
					.loading { text-align: center; padding: 20px; }
					.error { color: red; padding: 10px; }
					.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
				</style>
			</head>
			<body>
				<div class="header">
					<h1>Your Activities</h1>
					<button id="rename" class="btn">Rename Activities</button>
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
				</script>
			</body>
		</html>
	`, tokens.AccessToken, tokens.AccessToken)
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
