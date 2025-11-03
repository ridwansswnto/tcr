package agent

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

func EnsureRunnerBinary(dest, version string) error {
	if _, err := os.Stat(filepath.Join(dest, "config.sh")); err == nil {
		return nil
	}
	url := mapRuntimeToAsset(runtime.GOOS, runtime.GOARCH, version)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return ExtractTarGz(resp.Body, dest)
}

func mapRuntimeToAsset(goos, arch, version string) string {
	switch goos {
	case "linux":
		return fmt.Sprintf("https://github.com/actions/runner/releases/download/v%s/actions-runner-linux-x64-%s.tar.gz", version, version)
	case "darwin":
		return fmt.Sprintf("https://github.com/actions/runner/releases/download/v%s/actions-runner-osx-arm64-%s.tar.gz", version, version)
	default:
		return ""
	}
}
