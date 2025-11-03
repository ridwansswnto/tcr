package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/ridwandwisiswanto/tcr/internal/agent"
)

func main() {
	log.Println("🚀 Starting tower-agentd...")

	// 1️⃣ Load .env (biar TOWER_URL, GITHUB_OWNER, dsb kebaca)
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  .env not found or failed to load: %v", err)
	}

	// 2️⃣ Start agent
	a := agent.NewAgent()
	log.Printf("🌐 Tower URL: %s | Repo: %s", a.Config().TowerURL, a.Config().RepoFullName)

	if err := a.Run(); err != nil {
		log.Fatalf("❌ Agent exited: %v", err)
	}
}
