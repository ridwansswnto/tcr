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

// EnsureRunnerBinary memastikan binary GitHub Actions runner tersedia di /core, bukan di setiap runner instance.
func EnsureRunnerBinary(baseDir, version string) error {
	if baseDir == "" {
		return fmt.Errorf("empty base dir")
	}

	coreDir := filepath.Join(baseDir, "core")
	tarPath := filepath.Join(baseDir, fmt.Sprintf("runner-%s.tar.gz", version))
	configPath := filepath.Join(coreDir, "config.sh")

	// Kalau sudah ada, skip
	if _, err := os.Stat(configPath); err == nil {
		log.Printf("✅ Shared runner binary already exists at %s", coreDir)
		return nil
	}

	// Buat folder core
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		return fmt.Errorf("failed to create core dir: %v", err)
	}

	// Download binary tarball
	url := fmt.Sprintf("https://github.com/actions/runner/releases/download/v%s/actions-runner-linux-x64-%s.tar.gz", version, version)
	log.Printf("⬇️ Downloading GitHub runner v%s ...", version)

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

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save runner binary: %v", err)
	}

	// Ekstrak TAR ke coreDir
	log.Printf("📦 Extracting %s → %s", tarPath, coreDir)
	cmd := exec.Command("tar", "xzf", tarPath, "-C", coreDir, "--strip-components=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract runner: %v", err)
	}

	// Hapus archive setelah ekstraksi
	_ = os.Remove(tarPath)
	log.Printf("🧹 Removed archive %s after extraction", tarPath)

	log.Printf("✅ Runner binary v%s installed successfully at %s", version, coreDir)
	return nil
}
