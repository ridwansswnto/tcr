package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/ridwandwisiswanto/tcr/internal/github"
)

var (
	controllerURL = "http://localhost:8080"
	runnerID      = "runner-001"
	isBusy        = false
)

// Struct untuk menerima token dari Tower
type RegistrationPayload struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}

func main() {
	http.HandleFunc("/register-hybrid", func(w http.ResponseWriter, r *http.Request) {
		var payload github.RegistrationPayload
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &payload)

		if payload.Token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}

		runnerName := os.Getenv("RUNNER_NAME")
		if runnerName == "" {
			runnerName = "runner-001"
		}

		url := payload.URL
		log.Printf("🧩 Registering runner %s (Hybrid Mode)...", runnerName)

		if err := github.HybridRegister(payload.Token, url, runnerName); err != nil {
			log.Printf("❌ Registration failed: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}

		os.Setenv("LAST_RUNNER_TOKEN", payload.Token)
		go github.HybridRun()

		w.WriteHeader(http.StatusOK)
	})

	// Handle graceful shutdown (auto unregister)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		github.HybridUnregister()
		os.Exit(0)
	}()

	port := ":8081"
	log.Printf("🏃 Runner '%s' listening on %s", os.Getenv("RUNNER_NAME"), port)
	log.Fatal(http.ListenAndServe(port, nil))

	//if env := os.Getenv("RUNNER_ID"); env != "" {
	//	runnerID = env
	// }

	// go heartbeatLoop()
	// http.HandleFunc("/job", handleJob)
	// http.HandleFunc("/register", RegisterHandler)
	// http.HandleFunc("/register-api", RegisterAPIModeHandler)

	// port := ":8081"
	// log.Printf("🏃 Runner '%s' listening on %s", runnerID, port)
	// log.Fatal(http.ListenAndServe(port, nil))

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

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegistrationPayload
	body, _ := io.ReadAll(r.Body)
	json.Unmarshal(body, &payload)

	if payload.Token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	log.Printf("🪪 Received registration token, registering to GitHub Actions...")

	cmd := exec.Command(
		"./config.sh",
		"--url", payload.URL,
		"--token", payload.Token,
		"--name", os.Getenv("RUNNER_NAME"),
		"--unattended",
		"--replace",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Printf("❌ Failed to register runner: %v", err)
		http.Error(w, "runner registration failed", 500)
		return
	}

	log.Printf("✅ Runner successfully registered to GitHub Actions.")
	w.WriteHeader(http.StatusOK)
}

func RegisterAPIModeHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegistrationPayload
	body, _ := io.ReadAll(r.Body)
	json.Unmarshal(body, &payload)

	if payload.Token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")
	runnerName := os.Getenv("RUNNER_NAME")
	if runnerName == "" {
		runnerName = "runner-001"
	}

	log.Printf("🧩 Registering runner %s directly via GitHub API...", runnerName)
	if err := github.RegisterRunnerDirect(owner, repo, payload.Token, runnerName); err != nil {
		log.Printf("❌ Runner registration failed: %v", err)
		http.Error(w, fmt.Sprintf("failed: %v", err), 500)
		return
	}

	log.Printf("✅ Runner %s registered successfully via API!", runnerName)
	w.WriteHeader(http.StatusOK)
}
