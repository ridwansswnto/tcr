package github

import (
	"fmt"
	"net/http"
	"os"
)

// RemoveRunnerByID — hapus runner dari repo menggunakan REST API
func RemoveRunnerByID(runnerID int) error {
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	if token == "" || owner == "" || repo == "" {
		return fmt.Errorf("missing GITHUB_TOKEN, GITHUB_OWNER, or GITHUB_REPO")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runners/%d", owner, repo, runnerID)

	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call GitHub API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		return fmt.Errorf("GitHub API responded %d", resp.StatusCode)
	}
	return nil
}
