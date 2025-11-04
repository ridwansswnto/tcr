package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ridwandwisiswanto/tcr/internal/core"
)

var lastDispatchedID string

// StartDispatcher menjalankan goroutine yang terus memonitor queue
// dan mengirim job ke runner yang idle
func StartDispatcher() {
	go func() {
		for {
			time.Sleep(3 * time.Second)

			jobQueueMu.Lock()
			jobsSnapshot := make([]*core.Job, 0, len(jobQueue))
			for i := range jobQueue {
				jobsSnapshot = append(jobsSnapshot, &jobQueue[i])
			}
			jobQueueMu.Unlock()

			for _, job := range jobsSnapshot {
				if job.Status != "queued" {
					continue
				}

				// hindari mengirim job yang sama dua kali
				if job.ID == lastDispatchedID {
					continue
				}

				runner := GetIdleRunner()
				if runner == nil {
					break
				}

				jobQueueMu.Lock()
				if job.Status != "queued" {
					jobQueueMu.Unlock()
					continue
				}
				job.Status = "dispatched"
				jobQueueMu.Unlock()

				payload, _ := json.Marshal(job)
				url := fmt.Sprintf("http://%s:%s/job", runner.Address, runner.Port)
				resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
				if err != nil {
					log.Printf("❌ Failed to dispatch job %s to runner %s: %v", job.JobName, runner.ID, err)
					continue
				}
				resp.Body.Close()

				lastDispatchedID = job.ID
				log.Printf("🚀 Dispatched job '%s' (ID: %s) to runner '%s'", job.JobName, job.ID, runner.ID)
				MarkRunnerBusy(runner.ID, true)
			}
		}
	}()
}

// TriggerNextJob mencari job queued berikutnya segera tanpa menunggu interval loop
func TriggerNextJob() {
	go func() {
		jobQueueMu.Lock()
		var nextJob *core.Job
		for i := range jobQueue {
			if jobQueue[i].Status == "queued" {
				nextJob = &jobQueue[i]
				break
			}
		}
		jobQueueMu.Unlock()

		if nextJob == nil {
			return
		}

		runner := GetIdleRunner()
		if runner == nil {
			return
		}

		payload, _ := json.Marshal(nextJob)
		url := fmt.Sprintf("http://%s:%s/job", runner.Address, runner.Port)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("❌ Failed to trigger next job %s: %v", nextJob.JobName, err)
			return
		}
		resp.Body.Close()

		jobQueueMu.Lock()
		nextJob.Status = "dispatched"
		jobQueueMu.Unlock()

		lastDispatchedID = nextJob.ID
		log.Printf("⚡ Triggered next job '%s' (ID: %s) to runner '%s'", nextJob.JobName, nextJob.ID, runner.ID)
		MarkRunnerBusy(runner.ID, true)
	}()
}
