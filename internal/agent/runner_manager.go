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
	_ = os.RemoveAll(dir)
	_ = os.MkdirAll(dir, 0755)

	// Download binary langsung ke folder instance
	if err := EnsureRunnerBinary(dir, cfg.RunnerVersion); err != nil {
		return nil, fmt.Errorf("install binary: %w", err)
	}

	token, err := FetchTokenFromTower(cfg)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	configPath := filepath.Join(dir, "config.sh")
	cmd := exec.Command(configPath,
		"--unattended",
		"--url", fmt.Sprintf("https://github.com/%s", cfg.RepoFullName),
		"--token", token,
		"--name", name,
		"--replace",
	)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("config.sh failed: %w", err)
	}

	runCmd := exec.Command("./run.sh")
	runCmd.Dir = dir
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Start(); err != nil {
		return nil, fmt.Errorf("run.sh start failed: %w", err)
	}

	log.Printf("🏃 Runner %s started (dir=%s)", name, dir)
	return &Runner{ID: id, Name: name, Dir: dir, LastJobAt: time.Now()}, nil
}

// SpawnRunner membuat 1 instance runner baru berdasarkan shared core/
// func SpawnRunner(id int, cfg Config) (*Runner, error) {
// 	name := fmt.Sprintf("%s-agent-%02d", cfg.InstanceName, id)
// 	coreDir := filepath.Join(cfg.RunnerDir, "core")
// 	instanceDir := filepath.Join(cfg.RunnerDir, "instances", fmt.Sprintf("runner-%02d", id))

// 	// pastikan folder instance ada
// 	if err := os.MkdirAll(instanceDir, 0755); err != nil {
// 		return nil, fmt.Errorf("failed to create instance dir: %w", err)
// 	}

// 	// ambil token dari tower
// 	token, err := FetchTokenFromTower(cfg)
// 	if err != nil {
// 		return nil, fmt.Errorf("get token: %w", err)
// 	}

// 	// 🔗 gunakan coreDir sebagai template instance baru
// 	if err := CreateInstanceFromCore(coreDir, instanceDir); err != nil {
// 		return nil, fmt.Errorf("copy core files: %w", err)
// 	}

// 	// jalankan konfigurasi runner
// 	os.Remove(filepath.Join(instanceDir, ".runner"))
// 	configPath := filepath.Join(instanceDir, "config.sh")
// 	cmd := exec.Command(configPath,
// 		"--unattended",
// 		"--url", fmt.Sprintf("https://github.com/%s", cfg.RepoFullName),
// 		"--token", token,
// 		"--name", name,
// 		"--replace",
// 	)
// 	cmd.Dir = instanceDir
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	if err := cmd.Run(); err != nil {
// 		return nil, fmt.Errorf("config.sh failed: %w", err)
// 	}

// 	// 🚀 Jalankan runner-nya
// 	runCmd := exec.Command("./run.sh")
// 	runCmd.Dir = instanceDir
// 	runCmd.Stdout = os.Stdout
// 	runCmd.Stderr = os.Stderr
// 	if err := runCmd.Start(); err != nil {
// 		return nil, fmt.Errorf("run.sh start failed: %w", err)
// 	}

// 	log.Printf("🏃 Runner %s started (dir=%s)", name, instanceDir)
// 	return &Runner{
// 		ID:        id,
// 		Name:      name,
// 		Dir:       instanceDir,
// 		LastJobAt: time.Now(),
// 	}, nil
// }

// FetchTokenFromTower meminta token registrasi dari Tower
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
