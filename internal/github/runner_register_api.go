package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"time"
)

type RunnerRegisterRequest struct {
	Name          string `json:"name"`
	OSDescription string `json:"osDescription"`
	Version       string `json:"version"`
	Ephemeral     bool   `json:"ephemeral"`
}

type RunnerRegisterResponse struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	OSDescription string    `json:"osDescription"`
	Version       string    `json:"version"`
	Enabled       bool      `json:"enabled"`
	Status        string    `json:"status"`
	CreatedOn     time.Time `json:"createdOn"`
	LastOnline    time.Time `json:"lastOnline"`
	Busy          bool      `json:"busy"`
}

// RegisterRunnerDirect mendaftarkan runner langsung ke GitHub API tanpa config.sh
func RegisterRunnerDirect(owner, repo, token, runnerName string) error {
	// url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runners/registration-token", owner, repo)

	// Token ini hanya digunakan untuk registrasi sesi
	registerReq := RunnerRegisterRequest{
		Name:          runnerName,
		OSDescription: fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
		Version:       "2.317.0",
		Ephemeral:     false,
	}

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("https://pipelines.actions.githubusercontent.com/_apis/distributedtask/pools/1/agents?api-version=6.0-preview"),
		bytes.NewBuffer(body),
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("http request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed [%d]: %s", resp.StatusCode, string(respBody))
	}

	var result RunnerRegisterResponse
	json.NewDecoder(resp.Body).Decode(&result)

	log.Printf("✅ Runner %s registered to GitHub (ID: %d, Status: %s)", result.Name, result.ID, result.Status)

	// optional: simpan credential sementara
	SaveRunnerCredential(result, token)

	return nil
}
