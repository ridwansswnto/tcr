package agent

import (
	"log"
	"os"
	"os/exec"
	"time"
)

type Agent struct {
	runners []*Runner
	config  Config
}

func NewAgent() *Agent {
	cfg := LoadConfig()
	return &Agent{
		config:  cfg,
		runners: make([]*Runner, 0),
	}
}

func (a *Agent) Run() error {
	log.Printf("🧩 agentd started (max runners: %d)", a.config.MaxRunners)

	// 1. pastikan binary runner tersedia
	if err := EnsureRunnerBinary(a.config.RunnerDir, a.config.RunnerVersion); err != nil {
		return err
	}

	// 2. spawn runners
	for i := 1; i <= a.config.MaxRunners; i++ {
		r, err := SpawnRunner(i, a.config)
		if err != nil {
			log.Printf("⚠️ runner-%d failed spawn: %v", i, err)
			continue
		}
		a.runners = append(a.runners, r)
	}

	// 3. loop heartbeat
	go a.HeartbeatLoop()

	// 4. monitor idle
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

// DeregisterAll akan menghapus semua runner di VM ini dari GitHub
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
