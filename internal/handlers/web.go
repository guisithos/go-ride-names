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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
		<!DOCTYPE html>
		<html lang="pt-BR">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>zoAtleta - Seu treino, nossa piada!</title>
				
				<!-- Favicon -->
				<link rel="icon" type="image/png" sizes="32x32" href="/static/favicon/favicon-32x32.png">
				<link rel="icon" type="image/png" sizes="16x16" href="/static/favicon/favicon-16x16.png">
				<link rel="apple-touch-icon" sizes="180x180" href="/static/favicon/apple-touch-icon.png">
				<link rel="manifest" href="/static/site.webmanifest">
				<meta name="theme-color" content="#FC4C02">
				<style>
					body { 
						font-family: 'Segoe UI', Arial, sans-serif;
						max-width: 1000px;
						margin: 0 auto;
						padding: 20px;
						background-color: #f5f5f5;
						color: #333;
					}
					.header {
						display: flex;
						align-items: center;
						gap: 20px;
						margin-bottom: 40px;
					}
					.header img {
						height: 80px;
						width: auto;
					}
					.header-text {
						flex: 1;
					}
					h1 {
						font-size: 2.5em;
						margin: 0;
						color: #FC4C02;
					}
					.slogan {
						font-size: 1.2em;
						color: #666;
						margin: 5px 0;
					}
					.description {
						font-size: 1.1em;
						line-height: 1.6;
						margin: 30px 0;
						color: #444;
					}
					.features {
						background: white;
						padding: 25px;
						border-radius: 10px;
						box-shadow: 0 2px 5px rgba(0,0,0,0.1);
						margin: 30px 0;
					}
					.features h2 {
						color: #FC4C02;
						margin-top: 0;
					}
					.features ul {
						list-style-type: none;
						padding: 0;
					}
					.features li {
						margin: 15px 0;
						padding-left: 25px;
						position: relative;
					}
					.features li:before {
						content: "✓";
						position: absolute;
						left: 0;
						color: #FC4C02;
					}
					.connect-button {
						background-color: #FC4C02;
						color: white;
						padding: 15px 30px;
						border: none;
						border-radius: 5px;
						font-size: 1.1em;
						cursor: pointer;
						transition: background-color 0.3s;
						display: block;
						width: fit-content;
						margin: 30px auto;
						text-decoration: none;
					}
					.connect-button:hover {
						background-color: #E34402;
					}
					.strava-badge {
						display: block;
						margin: 20px auto;
						height: 40px;
						width: auto;
					}
					.container {
						background: white;
						padding: 40px;
						border-radius: 15px;
						box-shadow: 0 2px 10px rgba(0,0,0,0.1);
					}
				</style>
			</head>
			<body>
				<div class="container">
					<div class="header">
						<img src="/static/zoaAtleta_logo.png" alt="zoAtleta Logo">
						<div class="header-text">
							<h1>zoAtleta</h1>
							<div class="slogan">Seu treino, nossa piada</div>
						</div>
					</div>

					<div class="description">
						Aplicativo básico criado em Go que irá mudar o nome das suas atividades padrões do Strava para trocadilhos e piadas relacionadas ao esporte, com pitadas de séries, filmes, livros e um pouco de cultura nerd e geek.
					</div>

					<div class="features">
						<h2>Como funciona?</h2>
						<ul>
							<li>Conecte sua conta do Strava</li>
							<li>Suas atividades padrão serão renomeadas automaticamente</li>
							<li>Divirta-se com nomes criativos e engraçados</li>
							<li>Compartilhe com seus amigos</li>
						</ul>
					</div>

					<a href="/auth" class="connect-button">
						Conectar com Strava
					</a>

					<img src="/static/api_logo_cptblWith_strava_horiz_gray.png" 
						 alt="Compatible with Strava" 
						 class="strava-badge">
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html lang="pt-BR">
			<head>
				<meta charset="UTF-8">
				<meta name="viewport" content="width=device-width, initial-scale=1.0">
				<title>Dashboard - zoAtleta</title>
				
				<!-- Favicon -->
				<link rel="icon" type="image/png" sizes="32x32" href="/static/favicon/favicon-32x32.png">
				<link rel="icon" type="image/png" sizes="16x16" href="/static/favicon/favicon-16x16.png">
				<link rel="apple-touch-icon" sizes="180x180" href="/static/favicon/apple-touch-icon.png">
				<link rel="manifest" href="/static/site.webmanifest">
				<meta name="theme-color" content="#FC4C02">
				<style>
					body { 
						font-family: 'Segoe UI', Arial, sans-serif;
						max-width: 1200px;
						margin: 0 auto;
						padding: 20px;
						background-color: #f5f5f5;
						color: #333;
					}
					.header {
						background: white;
						padding: 20px;
						border-radius: 15px;
						box-shadow: 0 2px 10px rgba(0,0,0,0.1);
						display: flex;
						justify-content: space-between;
						align-items: center;
						margin-bottom: 30px;
					}
					.header-left {
						display: flex;
						align-items: center;
						gap: 20px;
					}
					.header img {
						height: 60px;
						width: auto;
					}
					.header-text h1 {
						font-size: 2em;
						margin: 0;
						color: #FC4C02;
					}
					.header-text .slogan {
						color: #666;
						font-size: 1em;
					}
					.buttons-container {
						display: flex;
						gap: 15px;
					}
					.btn {
						background: #FC4C02;
						color: white;
						padding: 12px 24px;
						border: none;
						border-radius: 5px;
						font-size: 1em;
						cursor: pointer;
						transition: background-color 0.3s;
						text-decoration: none;
						display: inline-flex;
						align-items: center;
						gap: 8px;
					}
					.btn:hover {
						background: #E34402;
					}
					.btn:disabled {
						background: #ccc;
						cursor: not-allowed;
					}
					.status {
						background: white;
						padding: 20px;
						border-radius: 15px;
						box-shadow: 0 2px 10px rgba(0,0,0,0.1);
						margin-bottom: 30px;
					}
					.status.active {
						border-left: 5px solid #2e7d32;
						background: #e8f5e9;
					}
					.status.inactive {
						border-left: 5px solid #c62828;
						background: #ffebee;
					}
					.activities-container {
						background: white;
						padding: 30px;
						border-radius: 15px;
						box-shadow: 0 2px 10px rgba(0,0,0,0.1);
					}
					.activity {
						border: 1px solid #eee;
						padding: 20px;
						margin: 15px 0;
						border-radius: 8px;
						transition: transform 0.2s;
					}
					.activity:hover {
						transform: translateX(5px);
						border-left: 5px solid #FC4C02;
					}
					.activity h3 {
						margin: 0 0 10px 0;
						color: #FC4C02;
					}
					.activity p {
						margin: 5px 0;
						color: #666;
					}
					.loading {
						text-align: center;
						padding: 40px;
						color: #666;
					}
					.error {
						color: #c62828;
						background: #ffebee;
						padding: 15px;
						border-radius: 8px;
						margin: 10px 0;
					}
					.footer {
						text-align: center;
						margin-top: 40px;
						padding: 20px;
						color: #666;
					}
					.footer img {
						height: 30px;
						margin: 10px;
					}
				</style>
			</head>
			<body>
				<div class="header">
					<div class="header-left">
						<img src="/static/zoaAtleta_logo.png" alt="zoAtleta Logo">
						<div class="header-text">
							<h1>zoAtleta</h1>
							<div class="slogan">Seu treino, nossa piada</div>
						</div>
					</div>
					<div class="buttons-container">
						<button id="rename" class="btn">
							<span>Renomear Todas</span>
						</button>
						<button id="subscribe" class="btn">
							<span>Ativar Auto-Renomeação</span>
						</button>
					</div>
				</div>

				<div id="subscription-status" class="status inactive">
					Auto-renomeação está atualmente inativa
				</div>

				<div class="activities-container">
					<h2>Suas Atividades</h2>
					<div id="activities">
						<div class="loading">Carregando atividades...</div>
					</div>
				</div>

				<div class="footer">
					<p>Conectado com</p>
					<img src="/static/api_logo_cptblWith_strava_horiz_gray.png" alt="Powered by Strava">
				</div>

				<script>
					// Function to format date
					function formatDate(dateStr) {
						const date = new Date(dateStr);
						return date.toLocaleDateString('pt-BR') + ' ' + date.toLocaleTimeString('pt-BR');
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
								throw new Error('Falha ao carregar atividades');
							}

							const activities = await response.json();
							const container = document.getElementById('activities');
							container.innerHTML = '';

							activities.forEach(activity => {
								const div = document.createElement('div');
								div.className = 'activity';
								div.innerHTML = `+"`"+`
									<h3>${activity.name}</h3>
									<p>Tipo: ${activity.type}</p>
									<p>Distância: ${formatDistance(activity.distance)}</p>
									<p>Data: ${formatDate(activity.start_date_local)}</p>
								`+"`"+`;
								container.appendChild(div);
							});
						} catch (error) {
							const container = document.getElementById('activities');
							container.innerHTML = '<div class="error">Erro ao carregar atividades: ' + error.message + '</div>';
						}
					}

					// Load activities when page loads
					loadActivities();

					// Handle rename button click
					document.getElementById('rename').addEventListener('click', async () => {
						const button = document.getElementById('rename');
						button.disabled = true;
						button.innerHTML = '<span>Renomeando...</span>';

						try {
							const response = await fetch('/rename-activities', {
								method: 'POST'
							});

							if (!response.ok) {
								throw new Error('Failed to rename activities');
							}

							// Reload activities after renaming
							await loadActivities();
						} catch (error) {
							console.error('Error:', error);
						} finally {
							button.disabled = false;
							button.innerHTML = '<span>Renomear Todas</span>';
						}
					});

					// Handle subscribe button click
					document.getElementById('subscribe').addEventListener('click', async () => {
						const button = document.getElementById('subscribe');
						button.disabled = true;
						button.innerHTML = '<span>Ativando...</span>';

						try {
							const response = await fetch('/subscribe', {
								method: 'POST'
							});

							if (!response.ok) {
								throw new Error('Failed to subscribe');
							}

							checkSubscriptionStatus();
						} catch (error) {
							console.error('Error:', error);
						} finally {
							button.disabled = false;
							button.innerHTML = '<span>Ativar Auto-Renomeação</span>';
						}
					});

					// Function to check subscription status
					async function checkSubscriptionStatus() {
						try {
							const response = await fetch('/subscription-status');
							const data = await response.json();
							
							const statusDiv = document.getElementById('subscription-status');
							if (data.active) {
								statusDiv.className = 'status active';
								statusDiv.textContent = 'Auto-renomeação está ativa';
							} else {
								statusDiv.className = 'status inactive';
								statusDiv.textContent = 'Auto-renomeação está inativa';
							}
						} catch (error) {
							console.error('Error checking status:', error);
						}
					}

					// Check status on page load
					checkSubscriptionStatus();
				</script>
			</body>
		</html>
	`, tokens.AccessToken)
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

	// Validate tokens
	if tokens.AccessToken == "" {
		log.Printf("Access token is empty")
		http.Error(w, "Invalid access token", http.StatusUnauthorized)
		return
	}

	// Get base URL from request or environment
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://" + r.Host
	}
	callbackURL := baseURL + "/webhook"
	log.Printf("Subscription attempt - Base URL: %s, Callback URL: %s", baseURL, callbackURL)

	// Validate Strava configuration
	if h.stravaConfig.StravaClientID == "" || h.stravaConfig.StravaClientSecret == "" {
		log.Printf("Error: Strava credentials not configured properly")
		http.Error(w, "Strava configuration error", http.StatusInternalServerError)
		return
	}

	verifyToken := os.Getenv("WEBHOOK_VERIFY_TOKEN")
	if verifyToken == "" {
		log.Printf("Error: WEBHOOK_VERIFY_TOKEN not configured")
		http.Error(w, "Webhook verify token not configured", http.StatusInternalServerError)
		return
	}

	log.Printf("Creating Strava client with ID: %s, Access Token: %s (first 10 chars)",
		h.stravaConfig.StravaClientID, tokens.AccessToken[:10])

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
	activityService := service.NewActivityService(client)

	err := activityService.SubscribeToWebhooks(callbackURL, verifyToken)
	if err != nil {
		log.Printf("Error managing webhook subscription: %v", err)
		http.Error(w, fmt.Sprintf("Error managing subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Store subscription status in session with longer expiration
	if err := h.sessions.Set("webhook_active", true); err != nil {
		log.Printf("Error storing webhook status: %v", err)
	}

	log.Printf("Webhook subscription is active")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": true})
}

func (h *WebHandler) handleSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokens, exists := h.sessions.GetTokens("user")
	if !exists {
		log.Printf("No tokens found in session")
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)

	// Check actual subscription status
	subs, err := client.ListWebhookSubscriptions()
	if err != nil {
		log.Printf("Error checking subscriptions: %v", err)
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	// Check if we have any active subscriptions
	active := len(subs) > 0
	if active {
		// Update session status
		if err := h.sessions.Set("webhook_active", true); err != nil {
			log.Printf("Error storing webhook status: %v", err)
		}
	} else {
		h.sessions.Set("webhook_active", nil)
	}

	log.Printf("Subscription status check - Active: %v, Subscriptions: %d", active, len(subs))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": active})
}
