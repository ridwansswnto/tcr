package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type JobResult struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	RunnerID  string    `json:"runner_id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ResultHandler menerima callback dari runner
func ResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var res JobResult
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	UpdateJobStatus(res.ID, res.Status)

	if res.Status == "success" || res.Status == "failed" {
		MarkRunnerBusy(res.RunnerID, false)
		UpdateJobStatus(res.ID, "done")
		log.Printf("✅ Job %s fully completed and marked done", res.ID)

		go TriggerNextJob() // langsung trigger job berikutnya
	}

	w.WriteHeader(http.StatusOK)
}

// RegisterResultRoute menambahkan route /job/result
func RegisterResultRoute() {
	http.HandleFunc("/job/result", ResultHandler)
}
