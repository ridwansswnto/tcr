package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EnsureRunnerBinary memastikan binary runner tersedia & tidak recursive
func EnsureRunnerBinary(baseDir, version string) error {
	if baseDir == "" {
		return fmt.Errorf("empty runner dir")
	}

	// Deteksi path recursive yang berbahaya
	if strings.Contains(baseDir, "runner-") {
		return fmt.Errorf("illegal nested runner path: %s", baseDir)
	}

	// Buat folder utama kalau belum ada
	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		log.Printf("⚠️ Runner base directory not found, creating: %s", baseDir)
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return fmt.Errorf("cannot create runner directory: %v", err)
		}
	}

	coreDir := filepath.Join(baseDir, "core")
	configPath := filepath.Join(coreDir, "config.sh")
	tarPath := filepath.Join(baseDir, fmt.Sprintf("runner-%s.tar.gz", version))

	// Jika binary sudah ada, skip
	if _, err := os.Stat(configPath); err == nil {
		log.Printf("✅ Runner binary already exists at %s", coreDir)
		return nil
	}

	// Unduh runner tarball
	url := fmt.Sprintf("https://github.com/actions/runner/releases/download/v%s/actions-runner-linux-x64-%s.tar.gz", version, version)
	log.Printf("⬇️ Downloading runner binary v%s from %s", version, url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download runner binary: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(tarPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save runner binary: %v", err)
	}

	// Ekstrak ke core/
	log.Printf("📦 Extracting %s to %s ...", tarPath, coreDir)
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		return fmt.Errorf("failed to create core dir: %v", err)
	}

	cmd := exec.Command("tar", "xzf", tarPath, "-C", coreDir, "--strip-components=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract runner: %v", err)
	}

	// Hapus tarball
	if err := os.Remove(tarPath); err == nil {
		log.Printf("🧹 Removed archive %s after extraction", tarPath)
	}

	// Pastikan config.sh bisa dieksekusi
	if err := os.Chmod(filepath.Join(coreDir, "config.sh"), 0755); err != nil {
		log.Printf("⚠️ Failed to chmod config.sh: %v", err)
	}

	log.Printf("✅ Runner binary v%s ready at %s", version, coreDir)
	return nil
}
