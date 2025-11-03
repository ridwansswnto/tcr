package github

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type RegistrationPayload struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

// HybridRegister menjalankan proses registrasi runner GitHub menggunakan binary resmi
func HybridRegister(token, url, runnerName string) error {
	runnerDir := os.Getenv("RUNNER_DIR")
	if runnerDir == "" {
		runnerDir = "./actions-runner"
	}

	// Pastikan folder runner sudah ada
	if _, err := os.Stat(filepath.Join(runnerDir, "config.sh")); os.IsNotExist(err) {
		return fmt.Errorf("config.sh not found in %s — download GitHub runner binary first", runnerDir)
	}

	log.Printf("🧩 Starting hybrid registration for runner %s ...", runnerName)

	cmd := exec.Command(
		"./config.sh",
		"--url", url,
		"--token", token,
		"--name", runnerName,
		"--unattended",
		"--replace",
	)
	cmd.Dir = runnerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("config.sh failed: %v", err)
	}

	log.Println("✅ Runner successfully registered to GitHub Actions.")
	return nil
}

// HybridRun menjalankan runner daemon-nya (run.sh)
func HybridRun() error {
	runnerDir := os.Getenv("RUNNER_DIR")
	if runnerDir == "" {
		runnerDir = "./actions-runner"
	}

	runScript := filepath.Join(runnerDir, "run.sh")
	if _, err := os.Stat(runScript); os.IsNotExist(err) {
		return fmt.Errorf("run.sh not found in %s", runnerDir)
	}

	cmd := exec.Command("./run.sh")
	cmd.Dir = runnerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("🏃 Starting GitHub Actions runner...")
	return cmd.Run()
}

// HybridUnregister membersihkan runner saat shutdown
func HybridUnregister() {
	runnerDir := os.Getenv("RUNNER_DIR")
	if runnerDir == "" {
		runnerDir = "./actions-runner"
	}

	cmd := exec.Command("./config.sh", "remove", "--token", os.Getenv("LAST_RUNNER_TOKEN"))
	cmd.Dir = runnerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("🧹 Unregistering runner from GitHub...")
	cmd.Run()
	log.Println("✅ Runner unregistered.")
}
