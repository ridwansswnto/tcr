package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	githubToken = os.Getenv("GITHUB_TOKEN")
	githubOwner = os.Getenv("GITHUB_OWNER")
	githubRepo  = os.Getenv("GITHUB_REPO")
	client      = &http.Client{Timeout: 10 * time.Second}
)

func apiURL(path string, q url.Values) string {
	base := fmt.Sprintf("https://api.github.com/repos/%s/%s", githubOwner, githubRepo)
	u := base + path
	if q != nil {
		u = u + "?" + q.Encode()
	}
	return u
}

// CountQueuedRuns returns number of workflow runs with status=queued
func CountQueuedRuns() (int, error) {
	q := url.Values{}
	q.Set("status", "queued")
	// per_page small to reduce payload; GitHub paginates — we'll only read first page count
	q.Set("per_page", "100")

	req, _ := http.NewRequest("GET", apiURL("/actions/workflow-runs", q), nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("github API %d", resp.StatusCode)
	}

	var data struct {
		TotalCount int `json:"total_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}
	return data.TotalCount, nil
}

// GetRunners returns all runners and count idle ones
func GetRunners() (total int, idle int, err error) {
	req, _ := http.NewRequest("GET", apiURL("/actions/runners", nil), nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return 0, 0, fmt.Errorf("github API %d", resp.StatusCode)
	}

	var data struct {
		TotalCount int `json:"total_count"`
		Runners    []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			Busy bool   `json:"busy"`
		} `json:"runners"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, err
	}

	total = data.TotalCount
	for _, r := range data.Runners {
		if !r.Busy {
			idle++
		}
	}
	return total, idle, nil
}

// GetRegistrationToken asks GitHub for registration token (for new runner)
func GetRegistrationToken() (string, time.Time, error) {
	req, _ := http.NewRequest("POST", apiURL("/actions/runners/registration-token", nil), nil)
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", time.Time{}, fmt.Errorf("github API %d", resp.StatusCode)
	}

	var data struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", time.Time{}, err
	}
	return data.Token, data.ExpiresAt, nil
}
