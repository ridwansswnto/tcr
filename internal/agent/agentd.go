package agent

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ridwandwisiswanto/tcr/internal/github"
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

		// Jika semua runner idle
		if AllRunnersIdle(a.runners, a.config.IdleTimeout) {
			log.Println("🧹 All runners idle — shutting down soon")
			a.DeregisterAll()

			// 🔒 Kosongkan daftar runner agar tidak loop terus
			a.runners = nil

			// 🚀 Kalau auto-shutdown aktif, hentikan VM
			if a.config.AutoShutdown {
				log.Println("💤 Auto-shutdown enabled, exiting agentd...")
				os.Exit(0)
			}

			// 🧘 Stop loop supaya gak spam deregister terus
			log.Println("🧘 All runners removed, stopping idle monitor loop.")
			break
		}
	}
}

// 🧹 DeregisterAll akan hapus semua runner di VM ini dari GitHub
func (a *Agent) DeregisterAll() {
	log.Println("🧹 Deregistering all runners (using GitHub API)")

	for _, r := range a.runners {
		// ambil ID dari nama (kita bisa simpan ID di struct Runner waktu spawn)
		runnerID, err := github.GetRunnerIDByName(r.Name)
		if err != nil {
			log.Printf("⚠️ Cannot find GitHub runner ID for %s: %v", r.Name, err)
			continue
		}

		if err := github.RemoveRunnerByID(runnerID); err != nil {
			log.Printf("❌ Failed to remove runner %s (id:%d): %v", r.Name, r.ID, err)
		} else {
			log.Printf("🗑 Runner %s (id:%d) removed successfully via GitHub API", r.Name, r.ID)
		}
	}
}

// ✅ Getter Config() untuk akses config dari luar package
func (a *Agent) Config() Config {
	return a.config
}
