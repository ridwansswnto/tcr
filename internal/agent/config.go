package agent

import (
	"os"
	"strconv"
)

type Config struct {
	TowerURL          string
	RunnerDir         string
	RepoFullName      string
	InstanceName      string
	MaxRunners        int
	HeartbeatInterval int
	IdleTimeout       int
	RunnerVersion     string
	AutoShutdown      bool
}

func LoadConfig() Config {
	return Config{
		TowerURL:          getEnv("TOWER_URL", "http://localhost:8080"),
		RunnerDir:         getEnv("RUNNER_DIR", "./actions-runner"),
		RepoFullName:      getEnv("GITHUB_OWNER", "user") + "/" + getEnv("GITHUB_REPO", "demo"),
		InstanceName:      getEnv("VM_NAME", "local-vm"),
		MaxRunners:        atoi(getEnv("MAX_RUNNERS_PER_VM", "5")),
		HeartbeatInterval: atoi(getEnv("HEARTBEAT_INTERVAL", "15")),
		IdleTimeout:       atoi(getEnv("VM_IDLE_TIMEOUT_SEC", "120")),
		RunnerVersion:     getEnv("GH_RUNNER_VERSION", "2.317.0"),
		AutoShutdown:      getEnv("AUTO_SHUTDOWN_ON_IDLE", "false") == "true",
	}
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
