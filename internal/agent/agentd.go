package agent

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Agent struct {
	runners []*Runner
	config  Config
}

func NewAgent() *Agent {
	cfg := LoadConfig()

	// ✅ Auto-fix runner path (absolute)
	if !strings.HasPrefix(cfg.RunnerDir, "/") {
		cwd, _ := os.Getwd()
		cfg.RunnerDir = filepath.Join(cwd, cfg.RunnerDir)
		log.Printf("📂 Fixed relative runner path → %s", cfg.RunnerDir)
	}

	// ✅ Pastikan folder ada (create jika belum)
	if _, err := os.Stat(cfg.RunnerDir); os.IsNotExist(err) {
		log.Printf("⚠️ Runner dir not found, creating at %s", cfg.RunnerDir)
		if err := os.MkdirAll(cfg.RunnerDir, 0755); err != nil {
			log.Fatalf("❌ Cannot create runner dir: %v", err)
		}
	}

	return &Agent{
		config:  cfg,
		runners: make([]*Runner, 0),
	}
}

func (a *Agent) Run() error {
	log.Printf("🚀 Starting tower-agentd...")
	log.Printf("🧩 agentd started (max runners: %d)", a.config.MaxRunners)

	// 1️⃣ Pastikan binary runner tersedia
	if err := EnsureRunnerBinary(a.config.RunnerDir, a.config.RunnerVersion); err != nil {
		log.Printf("⚠️ EnsureRunnerBinary failed: %v", err)
		return err
	}

	// 2️⃣ Spawn runners
	for i := 1; i <= a.config.MaxRunners; i++ {
		r, err := SpawnRunner(i, a.config)
		if err != nil {
			log.Printf("⚠️ runner-%d failed spawn: %v", i, err)
			continue
		}
		a.runners = append(a.runners, r)
	}

	// 3️⃣ Kirim heartbeat loop
	go a.HeartbeatLoop()

	// 4️⃣ Monitor idle
	a.MonitorIdle()

	return nil
}

func (a *Agent) MonitorIdle() {
	for {
		time.Sleep(15 * time.Second)
		if AllRunnersIdle(a.runners, a.config.IdleTimeout) {
			log.Println("🧹 All runners idle — shutting down soon")
			a.DeregisterAll()
			if a.config.AutoShutdown {
				os.Exit(0)
			}
		}
	}
}

// 🧹 DeregisterAll akan hapus semua runner di VM ini dari GitHub
func (a *Agent) DeregisterAll() {
	log.Println("🧹 Deregistering all runners (best-effort)")
	for _, r := range a.runners {
		removeCmd := exec.Command("./config.sh", "remove", "--unattended")
		removeCmd.Dir = r.Dir
		removeCmd.Stdout = os.Stdout
		removeCmd.Stderr = os.Stderr
		if err := removeCmd.Run(); err != nil {
			log.Printf("⚠️ Failed to remove runner %s: %v", r.Name, err)
		} else {
			log.Printf("🗑 Runner %s deregistered from GitHub", r.Name)
		}
	}
}
