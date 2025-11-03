package github

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ridwandwisiswanto/tcr/internal/controller"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	// Ambil secret dari ENV (sudah di-load di main.go lewat godotenv)
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		http.Error(w, "server misconfigured: missing GITHUB_WEBHOOK_SECRET", http.StatusInternalServerError)
		return
	}

	event := r.Header.Get("X-GitHub-Event")
	if event != "workflow_job" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := VerifySignature(r, secret); err != nil {
		log.Printf("❌ Invalid signature: %v", err)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var payload WorkflowJobPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	controller.AddJob(controller.Job{
		ID:        time.Now().Format("20060102150405"), // simple ID
		Action:    payload.Action,
		RepoOwner: payload.Repository.Owner.Login,
		RepoName:  payload.Repository.Name,
		JobName:   payload.WorkflowJob.Name,
		Status:    payload.WorkflowJob.Status,
		CreatedAt: time.Now(),
	})

	log.Printf("📦 Job queued: %s | repo: %s/%s | status: %s",
		payload.WorkflowJob.Name,
		payload.Repository.Owner.Login,
		payload.Repository.Name,
		payload.WorkflowJob.Status,
	)

	log.Printf("📦 Received job from repo %s/%s | action: %s | job: %s | status: %s",
		payload.Repository.Owner.Login,
		payload.Repository.Name,
		payload.Action,
		payload.WorkflowJob.Name,
		payload.WorkflowJob.Status,
	)

	w.WriteHeader(http.StatusOK)
}
