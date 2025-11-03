package main

import (
	"log"

	"github.com/ridwandwisiswanto/tcr/internal/agent"
)

func main() {
	log.Println("🚀 Starting tower-agentd...")
	a := agent.NewAgent()
	if err := a.Run(); err != nil {
		log.Fatalf("❌ Agent exited: %v", err)
	}
}
