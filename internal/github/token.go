package github

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type RunnerTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetRunnerRegistrationToken memanggil GitHub API untuk mendapatkan token runner
func GetRunnerRegistrationToken() (string, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	githubOwner := os.Getenv("GITHUB_OWNER")
	githubRepo := os.Getenv("GITHUB_REPO")

	if githubToken == "" || githubOwner == "" || githubRepo == "" {
		return "", fmt.Errorf("missing env vars: GITHUB_TOKEN, GITHUB_OWNER, or GITHUB_REPO")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runners/registration-token",
		githubOwner, githubRepo)

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("GitHub API call failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("GitHub API responded %d", resp.StatusCode)
	}

	var result RunnerTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	log.Printf("🔑 Received GitHub runner registration token (expires at %s)", result.ExpiresAt)
	return result.Token, nil
}
