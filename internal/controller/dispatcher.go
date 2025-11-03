package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var lastDispatchedID string
var dispatchLock sync.Mutex

// StartDispatcher memonitor queue dan mengirim job ke runner idle secara aman
func StartDispatcher() {
	go func() {
		dispatchLock.Lock()
		defer dispatchLock.Unlock()
		for {
			time.Sleep(500 * time.Millisecond)

			func() { // bungkus dalam fungsi supaya recover() bisa jalan
				defer func() {
					if r := recover(); r != nil {
						log.Printf("⚠️ Dispatcher recovered from panic: %v", r)
					}
				}()

				// ambil snapshot queue
				jobQueueMu.Lock()
				jobsSnapshot := make([]*Job, 0, len(jobQueue))
				for i := range jobQueue {
					jobsSnapshot = append(jobsSnapshot, &jobQueue[i])
				}
				jobQueueMu.Unlock()

				for _, job := range jobsSnapshot {
					// skip job yang tidak perlu dikirim
					if job.Status != "queued" || job.ID == lastDispatchedID {
						continue
					}

					runner := GetIdleRunner()
					if runner == nil {
						break
					}

					// ubah status sebelum kirim (atomic)
					jobQueueMu.Lock()
					if job.Status != "queued" {
						jobQueueMu.Unlock()
						continue
					}
					job.Status = "dispatched"
					jobQueueMu.Unlock()

					// ✅ Tandai runner busy SEBELUM mengirim
					MarkRunnerBusy(runner.ID, true)

					go func(j *Job, r *Runner) {
						defer func() {
							if r := recover(); r != nil {
								log.Printf("⚠️ Dispatch goroutine recovered: %v", r)
							}
						}()

						payload, _ := json.Marshal(j)
						url := fmt.Sprintf("http://%s:%s/job", r.Address, r.Port)

						// Tambahkan jeda kecil agar runner sempat idle sebelum dikirim job baru
						time.Sleep(1 * time.Second)

						client := &http.Client{Timeout: 8 * time.Second}

						resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
						if err != nil {
							log.Printf("❌ Failed to dispatch job %s to runner %s: %v", j.JobName, r.ID, err)
							return
						}
						resp.Body.Close()

						lastDispatchedID = j.ID
						log.Printf("🚀 Dispatched job '%s' (ID: %s) to runner '%s'", j.JobName, j.ID, r.ID)
						MarkRunnerBusy(r.ID, true)
					}(job, runner)
				}
			}()
		}
	}()
}

// TriggerNextJob langsung mencari job queued berikutnya (dipanggil dari ResultHandler)
func TriggerNextJob() {
	go func() {
		dispatchLock.Lock()
		defer dispatchLock.Unlock()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️ TriggerNextJob recovered from panic: %v", r)
			}
		}()

		jobQueueMu.Lock()
		var nextJob *Job
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

		// jeda kecil agar runner siap
		//time.Sleep(1 * time.Second)

		client := &http.Client{Timeout: 8 * time.Second}
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("❌ Failed to trigger next job %s: %v", nextJob.JobName, err)
			return
		}
		defer resp.Body.Close()

		// pastikan sukses
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			jobQueueMu.Lock()
			nextJob.Status = "dispatched"
			jobQueueMu.Unlock()

			lastDispatchedID = nextJob.ID
			log.Printf("⚡ Triggered next job '%s' (ID: %s) to runner '%s'", nextJob.JobName, nextJob.ID, runner.ID)
			MarkRunnerBusy(runner.ID, true)
		} else {
			log.Printf("⚠️ Runner %s failed to accept job %s (HTTP %d)", runner.ID, nextJob.JobName, resp.StatusCode)
		}

		jobQueueMu.Lock()
		nextJob.Status = "dispatched"
		jobQueueMu.Unlock()

		lastDispatchedID = nextJob.ID
		log.Printf("⚡ Triggered next job '%s' (ID: %s) to runner '%s'", nextJob.JobName, nextJob.ID, runner.ID)
		MarkRunnerBusy(runner.ID, true)
	}()
}
