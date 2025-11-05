package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/ridwandwisiswanto/tcr/internal/github"
)

var (
	pollIntervalSec  int
	maxScaleStep     int
	globalMaxRunners int
	spawnMethod      string
	agentRegisterURL string
	gcloudProject    string
	gcpMigName       string
	mode             string
)

func atoiEnv(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func StartPoller() {
	mode = os.Getenv("MODE")
	pollIntervalSec = atoiEnv("POLL_INTERVAL_SECONDS", 30)
	maxScaleStep = atoiEnv("SCALE_STEP_MAX", 3)
	globalMaxRunners = atoiEnv("MAX_RUNNERS_TOTAL", 20)
	spawnMethod = os.Getenv("SPAWN_METHOD")
	agentRegisterURL = getEnv("AGENT_REGISTRATION_ENDPOINT", "http://localhost:8081/register-hybrid")
	gcloudProject = os.Getenv("GCP_PROJECT")
	gcpMigName = os.Getenv("GCP_MIG_NAME")
	log.Printf("🧩 MODE env detected = '%s'", mode)

	if mode != "polling" {
		log.Printf("🔕 Poller disabled (MODE=%s)", mode)
		return
	}
	log.Printf("🔁 Starting polling engine (interval=%ds)", pollIntervalSec)
	go pollLoop()
}

func pollLoop() {
	ticker := time.NewTicker(time.Duration(pollIntervalSec) * time.Second)
	defer ticker.Stop()

	var backoff time.Duration
	for {
		select {
		case <-ticker.C:
			backoff = 0 // reset on each scheduled tick
			processOnce()
		default:
			if backoff > 0 {
				time.Sleep(backoff)
				// exponential backoff clamp
				if backoff < 60*time.Second {
					backoff *= 2
				}
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func processOnce() {
	queued, err := github.CountQueuedRuns()
	if err != nil {
		log.Printf("⚠️ poll error (queued): %v", err)
		return
	}

	total, idle, err := github.GetRunners()
	if err != nil {
		log.Printf("⚠️ poll error (runners): %v", err)
		return
	}

	log.Printf("📡 Poll: queued=%d | total_runners=%d | idle=%d", queued, total, idle)

	// compute how many more runners to spawn
	need := queued - idle
	if need <= 0 {
		// maybe scale down idle runners if idle > queued substantially
		// Strategy: keep min( idle-queued, some threshold)
		trim := idle - queued
		if trim > 0 {
			// Do not remove below zero, and respect global cap
			toRemove := trim
			if toRemove > maxScaleStep {
				toRemove = maxScaleStep
			}
			if toRemove > 0 {
				go scaleDown(toRemove)
			}
		}
		return
	}

	// respect global max runners
	remainingCapacity := globalMaxRunners - total
	if remainingCapacity <= 0 {
		log.Printf("⚠️ cannot scale up: reached global max %d", globalMaxRunners)
		return
	}
	if need > remainingCapacity {
		need = remainingCapacity
	}
	if need > maxScaleStep {
		need = maxScaleStep
	}

	log.Printf("🧩 Scaling up: need=%d", need)
	if spawnMethod == "gcp_mig" {
		if err := scaleUpViaGCP(need); err != nil {
			log.Printf("❌ scaleUpViaGCP failed: %v", err)
		}
	} else {
		if err := spawnLocalRunners(need); err != nil {
			log.Printf("❌ spawnLocalRunners failed: %v", err)
		}
	}
}

// spawnLocalRunners — instruct existing agent manager / launcher to create new runner instances
// Implementation: Tower triggers its agent registration flow (call internal agent endpoint) or calls local scripts.
// Here we call Tower internal endpoint /admin/agent/register which should instruct agent to create runners on VMs.
func spawnLocalRunners(n int) error {
	for i := 0; i < n; i++ {
		payload := map[string]string{"action": "spawn"}
		b, _ := json.Marshal(payload)
		resp, err := httpPost(agentRegisterURL, "application/json", b)
		if err != nil {
			return err
		}
		resp.Body.Close()
		time.Sleep(500 * time.Millisecond) // small spacing
	}
	return nil
}

func httpPost(url, contentType string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", contentType)
	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

// scaleUpViaGCP uses gcloud CLI as a quick placeholder (recommended: replace with GCP Compute API)
func scaleUpViaGCP(n int) error {
	if gcloudProject == "" || gcpMigName == "" {
		return fmt.Errorf("gcloud project or MIG not set")
	}
	// Example: gcloud compute instance-groups managed resize NAME --size=NEW_SIZE --project=PROJECT
	// This is a placeholder: you must compute new size (current+ n) and call CLI
	cmd := exec.Command("gcloud",
		"compute", "instance-groups", "managed", "resize",
		gcpMigName,
		"--project", gcloudProject,
		"--zone", "asia-southeast1-a",
		"--size", fmt.Sprintf("%d", n),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gcloud resize failed: %v %s", err, string(out))
	}
	return nil
}

func scaleDown(n int) {
	// simple placeholder: remove idle runners via GitHub API (we expect agentd to detect VM idle and shutdown)
	// Option: call GitHub to list idle runners and delete first n via API
	log.Printf("🧹 scaleDown requested: %d", n)
	// TODO: implement deletion logic with caution and safety (drain, unregister)
}
