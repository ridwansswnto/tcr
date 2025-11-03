package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

// EnsureRunnerBinary memastikan binary runner tersedia di direktori yang benar.
func EnsureRunnerBinary(dir, version string) error {
	// Pastikan direktori runner ada
	if dir == "" {
		return fmt.Errorf("empty runner dir")
	}

	// Jika folder belum ada, buat
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("⚠️ Runner directory not found, creating: %s", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("cannot create runner directory: %v", err)
		}
	}

	// Cek apakah config.sh sudah ada
	configPath := filepath.Join(dir, "config.sh")
	if _, err := os.Stat(configPath); err == nil {
		log.Printf("✅ Runner binary already exists at %s", dir)
		return nil
	}

	// Kalau belum ada, auto-download
	log.Printf("⬇️ Runner binary not found, downloading v%s...", version)

	url := fmt.Sprintf("https://github.com/actions/runner/releases/download/v%s/actions-runner-linux-x64-%s.tar.gz", version, version)
	tmpPath := filepath.Join(dir, fmt.Sprintf("runner-%s.tar.gz", version))

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download runner binary: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save runner binary: %v", err)
	}

	// Extract file tar.gz
	log.Printf("📦 Extracting %s ...", tmpPath)
	cmd := exec.Command("tar", "xzf", tmpPath, "-C", dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract runner: %v", err)
	}

	os.Remove(tmpPath)
	log.Printf("✅ Runner binary downloaded & extracted successfully to %s", dir)
	return nil
}
