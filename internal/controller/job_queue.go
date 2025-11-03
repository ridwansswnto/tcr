package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type Job struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	RepoOwner string    `json:"repo_owner"`
	RepoName  string    `json:"repo_name"`
	JobName   string    `json:"job_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	jobQueue   = []Job{}
	jobQueueMu sync.Mutex
)

// AddJob menambahkan job baru ke queue
func AddJob(j Job) {
	jobQueueMu.Lock()
	defer jobQueueMu.Unlock()

	jobQueue = append(jobQueue, j)
	// jobQueueGauge.Set(float64(len(jobQueue))) // 🟢 metrics update

	log.Printf("🧩 Job added to queue: %s (%s/%s)", j.JobName, j.RepoOwner, j.RepoName)
}

// GetJobs mengembalikan semua job yang ada di queue
func GetJobs() []Job {
	jobQueueMu.Lock()
	defer jobQueueMu.Unlock()

	// copy agar thread-safe
	jobs := make([]Job, len(jobQueue))
	copy(jobs, jobQueue)
	return jobs
}

// UpdateJobStatus memperbarui status job di queue berdasarkan ID
func UpdateJobStatus(id string, status string) {
	jobQueueMu.Lock()
	defer jobQueueMu.Unlock()

	for i := range jobQueue {
		if jobQueue[i].ID == id {
			jobQueue[i].Status = status
			log.Printf("🟡 Job %s status updated to %s", id, status)
			return
		}
	}

	log.Printf("⚠️ Job %s not found for status update", id)
}

// RegisterHTTPRoutes menambahkan route HTTP /jobs
func RegisterHTTPRoutes() {
	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		jobs := GetJobs()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jobs)
	})
}
