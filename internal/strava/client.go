package strava

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	baseURL                = "https://www.strava.com/api/v3"
	authURL                = "https://www.strava.com/oauth/token"
	activitiesURL          = baseURL + "/athlete/activities"
	webhookSubscriptionURL = baseURL + "/push_subscriptions"
)

type Client struct {
	accessToken  string
	refreshToken string
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

type TokenResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type UpdateActivityRequest struct {
	Name string `json:"name"`
}

type WebhookSubscription struct {
	ID            int64  `json:"id"`
	ApplicationID int    `json:"application_id"`
	CallbackURL   string `json:"callback_url"`
	VerifyToken   string `json:"verify_token"`
}

func NewClient(accessToken, refreshToken, clientID, clientSecret string) *Client {
	return &Client{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
	}
}

func (c *Client) RefreshToken() (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("refresh_token", c.refreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := c.httpClient.PostForm(authURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to refresh token: status=%d, body=%s", resp.StatusCode, body)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %v", err)
	}

	// Update client's tokens
	c.accessToken = tokenResp.AccessToken
	c.refreshToken = tokenResp.RefreshToken

	return &tokenResp, nil
}

// handle automatic token refresh
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	// Add authorization header
	if c.accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Handle token refresh if needed
	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("Token expired, attempting refresh")
		newTokens, err := c.RefreshToken()
		if err != nil {
			return nil, fmt.Errorf("token refresh failed: %v", err)
		}

		// Update tokens and retry request
		c.accessToken = newTokens.AccessToken
		c.refreshToken = newTokens.RefreshToken
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
		return c.httpClient.Do(req)
	}

	return resp, nil
}

// Update existing methods to use doRequest
func (c *Client) GetAuthenticatedAthlete() (*Athlete, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/athlete", baseURL), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var athlete Athlete
	if err := json.NewDecoder(resp.Body).Decode(&athlete); err != nil {
		return nil, err
	}

	return &athlete, nil
}

func (c *Client) GetAthleteActivities(page, perPage int, before, after int64) ([]Activity, error) {
	// Build query parameters
	query := url.Values{}
	query.Add("page", fmt.Sprintf("%d", page))
	query.Add("per_page", fmt.Sprintf("%d", perPage))

	if before != 0 {
		query.Add("before", fmt.Sprintf("%d", before))
	}
	if after != 0 {
		query.Add("after", fmt.Sprintf("%d", after))
	}

	// Create request
	req, err := http.NewRequest("GET", activitiesURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add authorization header
	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	// Check for successful status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s", string(body))
	}

	// Parse response
	var activities []Activity
	if err := json.Unmarshal(body, &activities); err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}

	return activities, nil
}

func (c *Client) UpdateActivity(activityID int64, name string) error {
	updateURL := fmt.Sprintf("%s/activities/%d", baseURL, activityID)

	// Create request body
	reqBody := UpdateActivityRequest{
		Name: name,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("error marshaling request: %v", err)
	}

	// Create request
	req, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	// Make the request
	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update activity: status=%d, body=%s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) CreateWebhookSubscription(callbackURL, verifyToken string) (*WebhookSubscription, error) {
	// Always delete ALL existing subscriptions first
	subs, err := c.ListWebhookSubscriptions()
	if err == nil {
		for _, sub := range subs {
			log.Printf("Deleting existing subscription ID: %d (URL: %s)", sub.ID, sub.CallbackURL)
			err := c.DeleteWebhookSubscription(sub.ID)
			if err != nil {
				log.Printf("Warning: failed to delete subscription %d: %v", sub.ID, err)
			}
		}
	}

	// Wait a moment to ensure deletion is processed
	time.Sleep(time.Second)

	// Now create the new subscription
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("callback_url", callbackURL)
	data.Set("verify_token", verifyToken)

	req, err := http.NewRequest("POST", webhookSubscriptionURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Create subscription response: Status=%d, Body=%s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create subscription: status=%d, body=%s",
			resp.StatusCode, string(body))
	}

	var subscription WebhookSubscription
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&subscription); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &subscription, nil
}

func (c *Client) GetActivity(activityID int64) (*Activity, error) {
	url := fmt.Sprintf("%s/activities/%d", baseURL, activityID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get activity: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var activity Activity
	if err := json.NewDecoder(resp.Body).Decode(&activity); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return &activity, nil
}

func (c *Client) ListWebhookSubscriptions() ([]WebhookSubscription, error) {
	// Build URL with query parameters
	u, err := url.Parse(webhookSubscriptionURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("client_id", c.clientID)
	q.Set("client_secret", c.clientSecret)
	u.RawQuery = q.Encode()

	log.Printf("Listing webhook subscriptions - URL: %s", u.String())

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Log the raw response
	body, _ := io.ReadAll(resp.Body)
	log.Printf("List subscriptions response: Status=%d, Body=%s", resp.StatusCode, string(body))

	// Create new reader for JSON decoder
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	var subscriptions []WebhookSubscription
	if err := json.NewDecoder(resp.Body).Decode(&subscriptions); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return subscriptions, nil
}

func (c *Client) GetActivities() ([]Activity, error) {
	req, err := http.NewRequest("GET", activitiesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var activities []Activity
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	return activities, nil
}

func (c *Client) DeleteWebhookSubscription(subscriptionID int64) error {
	// Build the URL with query parameters instead of form data
	deleteURL := fmt.Sprintf("%s/%d?client_id=%s&client_secret=%s",
		webhookSubscriptionURL,
		subscriptionID,
		c.clientID,
		c.clientSecret)

	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body for error reporting
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Delete subscription response: Status=%d, Body=%s", resp.StatusCode, string(body))

	// Consider both 204 (success) and 404 (already deleted) as success
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete subscription: status=%d, body=%s",
			resp.StatusCode, string(body))
	}

	return nil
}
