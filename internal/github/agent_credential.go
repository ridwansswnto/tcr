package github

import (
	"encoding/json"
	"log"
	"os"
)

type RunnerCredential struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Token     string `json:"token"`
	Timestamp string `json:"timestamp"`
}

// SaveRunnerCredential menyimpan credential runner ke file lokal
func SaveRunnerCredential(runner any, token string) {
	file := ".tcr_runner_credential.json"
	data := map[string]any{
		"runner": runner,
		"token":  token,
	}

	bytes, _ := json.MarshalIndent(data, "", "  ")
	if err := os.WriteFile(file, bytes, 0644); err != nil {
		log.Printf("⚠️ Failed to save credential: %v", err)
		return
	}

	log.Printf("💾 Saved runner credential to %s", file)
}
