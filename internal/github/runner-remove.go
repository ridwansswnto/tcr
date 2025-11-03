package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// RemoveRunnerByID — hapus runner dari repo menggunakan REST API
func GetRunnerIDByName(name string) (int, error) {
	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runners", owner, repo)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to query runners: %v", err)
	}
	defer resp.Body.Close()

	var data struct {
		Runners []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"runners"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, fmt.Errorf("decode: %v", err)
	}

	for _, r := range data.Runners {
		if r.Name == name {
			return r.ID, nil
		}
	}

	return 0, fmt.Errorf("runner %s not found", name)
}
