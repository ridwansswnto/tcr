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

// SpawnRunner membuat 1 instance runner baru berdasarkan binary di core/
func SpawnRunner(id int, cfg Config) (*Runner, error) {
	name := fmt.Sprintf("%s-agent-%02d", cfg.InstanceName, id)

	// Struktur direktori:
	// /opt/tcr/actions-runner/
	// ├── core/        -> berisi binary (config.sh, run.sh, dsb)
	// └── instances/
	//     ├── runner-01/
	//     ├── runner-02/
	instanceDir := filepath.Join(cfg.RunnerDir, "instances", fmt.Sprintf("runner-%02d", id))
	coreDir := filepath.Join(cfg.RunnerDir, "core")

	// Pastikan instance folder ada
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir instance: %w", err)
	}

	// Ambil token dari Tower
	token, err := FetchTokenFromTower(cfg)
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	// Jalankan config.sh dari core, tapi dengan working dir di instanceDir
	configPath := filepath.Join(coreDir, "config.sh")
	cmd := exec.Command(configPath,
		"--unattended",
		"--url", fmt.Sprintf("https://github.com/%s", cfg.RepoFullName),
		"--token", token,
		"--name", name,
		"--replace",
	)
	cmd.Dir = instanceDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("⚙️  Registering runner %s ...", name)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("config.sh failed: %w", err)
	}

	// Jalankan run.sh juga dari core (tapi workdir = instance)
	runPath := filepath.Join(coreDir, "run.sh")
	go func() {
		runCmd := exec.Command(runPath)
		runCmd.Dir = instanceDir
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr
		if err := runCmd.Run(); err != nil {
			log.Printf("⚠️ runner-%d stopped: %v", id, err)
		}
	}()

	log.Printf("🏃 Runner %s started (dir=%s)", name, instanceDir)
	return &Runner{
		ID:        id,
		Name:      name,
		Dir:       instanceDir,
		LastJobAt: time.Now(),
	}, nil
}

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
