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

// EnsureRunnerBinary memastikan binary core runner tersedia di <baseDir>/core.
func EnsureRunnerBinary(baseDir, version string) error {
	if baseDir == "" {
		return fmt.Errorf("empty base dir")
	}

	coreDir := filepath.Join(baseDir, "core")
	tarPath := filepath.Join(baseDir, fmt.Sprintf("runner-%s.tar.gz", version))
	configPath := filepath.Join(coreDir, "config.sh")

	// ✅ 1. Cek apakah binary core sudah tersedia
	if _, err := os.Stat(configPath); err == nil {
		log.Printf("✅ Shared runner binary already exists at %s", coreDir)
		return nil
	}

	// ✅ 2. Pastikan folder core bersih dan ada
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		return fmt.Errorf("failed to create core dir: %v", err)
	}

	// Hapus sisa-sisa file lama kalau ada (biar gak nested recursive)
	os.RemoveAll(filepath.Join(coreDir, "core"))
	os.RemoveAll(filepath.Join(coreDir, "instances"))

	// ✅ 3. Download binary tarball dari GitHub
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

	// ✅ 4. Ekstrak tarball ke folder core
	log.Printf("📦 Extracting %s → %s", tarPath, coreDir)
	cmd := exec.Command("tar", "xzf", tarPath, "-C", coreDir, "--strip-components=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract runner: %v", err)
	}

	// ✅ 5. Hapus file archive setelah sukses ekstraksi
	_ = os.Remove(tarPath)
	log.Printf("🧹 Removed archive %s after extraction", tarPath)
	log.Printf("✅ Runner binary v%s installed successfully at %s", version, coreDir)

	return nil
}

