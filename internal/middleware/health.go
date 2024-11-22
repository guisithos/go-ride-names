package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

type HealthResponse struct {
	Status      string                 `json:"status"`
	Timestamp   string                 `json:"timestamp"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Checks      map[string]CheckStatus `json:"checks"`
	System      SystemInfo             `json:"system"`
}

type CheckStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type SystemInfo struct {
	NumGoroutines int    `json:"num_goroutines"`
	NumCPU        int    `json:"num_cpu"`
	HeapInUse     uint64 `json:"heap_in_use"`
}

// Health middleware checks
func Health(version, env string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				checks := performHealthChecks()
				status := "ok"

				// If any check failed, mark as error
				for _, check := range checks {
					if check.Status != "ok" {
						status = "error"
						break
					}
				}

				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)

				response := HealthResponse{
					Status:      status,
					Timestamp:   time.Now().UTC().Format(time.RFC3339),
					Version:     version,
					Environment: env,
					Checks:      checks,
					System: SystemInfo{
						NumGoroutines: runtime.NumGoroutine(),
						NumCPU:        runtime.NumCPU(),
						HeapInUse:     memStats.HeapInuse,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				if status != "ok" {
					w.WriteHeader(http.StatusServiceUnavailable)
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func performHealthChecks() map[string]CheckStatus {
	checks := make(map[string]CheckStatus)

	// Check Strava API health by attempting to make a request
	stravaResp, err := http.Get("https://www.strava.com/api/v3/athlete")
	if err != nil {
		checks["strava_api"] = CheckStatus{
			Status:  "error",
			Message: fmt.Sprintf("Failed to connect to Strava API: %v", err),
		}
	} else {
		defer stravaResp.Body.Close()

		// Check if the API is responding with expected status codes
		// 401 is expected without auth token, which means API is up
		if stravaResp.StatusCode == http.StatusUnauthorized {
			checks["strava_api"] = CheckStatus{
				Status:  "ok",
				Message: "Strava API responding normally",
			}
		} else {
			checks["strava_api"] = CheckStatus{
				Status:  "error",
				Message: fmt.Sprintf("Unexpected Strava API response: %d", stravaResp.StatusCode),
			}
		}
	}

	return checks
}
