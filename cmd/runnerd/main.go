package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	controllerURL = "http://localhost:8080"
	runnerID      = "runner-001"
	isBusy        = false
)

func main() {
	if env := os.Getenv("RUNNER_ID"); env != "" {
		runnerID = env
	}

	go heartbeatLoop()
	http.HandleFunc("/job", handleJob)

	port := ":8081"
	log.Printf("🏃 Runner '%s' listening on %s", runnerID, port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func heartbeatLoop() {
	for {
		time.Sleep(10 * time.Second)
		if isBusy {
			continue
		}
		url := fmt.Sprintf("%s/heartbeat?id=%s&port=8081", controllerURL, runnerID)
		_, err := http.Get(url)
		if err != nil {
			log.Printf("⚠️ Heartbeat failed: %v", err)
		}
	}
}

func handleJob(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var job map[string]interface{}
	json.Unmarshal(body, &job)

	jobID := fmt.Sprintf("%v", job["id"])
	log.Printf("📦 Received job: %v", job)

	isBusy = true
	reportResult(jobID, "running")

	time.Sleep(5 * time.Second) // simulate work

	// misal 90% success, 10% fail (random)
	if time.Now().Unix()%10 == 0 {
		reportResult(jobID, "failed")
		log.Printf("❌ Job %s failed", jobID)
	} else {
		reportResult(jobID, "success")
		log.Printf("✅ Job %s done", jobID)
	}

	isBusy = false
	w.WriteHeader(http.StatusOK)
}

func reportResult(jobID, status string) {
	body := fmt.Sprintf(`{"id":"%s","status":"%s","runner_id":"%s"}`, jobID, status, runnerID)
	url := fmt.Sprintf("%s/job/result", controllerURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(body)))
	if err != nil {
		log.Printf("⚠️ Failed to report result: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("📨 Reported job %s as %s", jobID, status)
}
