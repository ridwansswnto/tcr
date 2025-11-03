package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Runner struct {
	ID        int
	Name      string
	Dir       string
	LastJobAt time.Time
}

func SpawnRunner(id int, cfg Config) (*Runner, error) {
	name := fmt.Sprintf("%s-agent-%02d", cfg.InstanceName, id)
	dir := filepath.Join(cfg.RunnerDir, fmt.Sprintf("runner-%02d", id))
	os.MkdirAll(dir, 0755)

	token, err := FetchTokenFromTower(cfg)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	// copy binary skeleton
	if err := CopyDirContents(cfg.RunnerDir, dir); err != nil {
		return nil, fmt.Errorf("copy: %w", err)
	}

	cmd := exec.Command("./config.sh", "--unattended",
		"--url", fmt.Sprintf("https://github.com/%s", cfg.RepoFullName),
		"--token", token, "--name", name, "--replace")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("config.sh: %w", err)
	}

	go exec.Command("./run.sh").Start()
	log.Printf("🏃 runner-%d started", id)
	return &Runner{ID: id, Name: name, Dir: dir, LastJobAt: time.Now()}, nil
}

func FetchTokenFromTower(cfg Config) (string, error) {
	url := fmt.Sprintf("%s/github/token", cfg.TowerURL)
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		return "", fmt.Errorf("failed fetch token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tower responded %d: %s", resp.StatusCode, string(body))
	}

	var data struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	if data.Token == "" {
		return "", fmt.Errorf("empty token from tower")
	}
	return data.Token, nil
}

// AllRunnersIdle memeriksa apakah semua runner idle dalam durasi tertentu
func AllRunnersIdle(runners []*Runner, idleTimeout int) bool {
	if len(runners) == 0 {
		return true
	}
	for _, r := range runners {
		if time.Since(r.LastJobAt) < time.Duration(idleTimeout)*time.Second {
			return false
		}
	}
	return true
}
