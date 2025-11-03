package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/ridwandwisiswanto/tcr/internal/controller"
	"github.com/ridwandwisiswanto/tcr/internal/github"
)

func main() {
	_ = godotenv.Load()

	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("❌ GITHUB_WEBHOOK_SECRET is not set.")
	}

	http.HandleFunc("/register-runner", func(w http.ResponseWriter, r *http.Request) {
		token, err := github.GetRunnerRegistrationToken()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "GitHub registration token: %s\n", token)
	})

	http.HandleFunc("/register-to-runner", func(w http.ResponseWriter, r *http.Request) {
		token, err := github.GetRunnerRegistrationToken()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		payload := map[string]string{
			"token": token,
			"url":   fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO")),
		}

		data, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:8081/register", "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("❌ Failed to send registration token to runner: %v", err)
			http.Error(w, "failed to send token", 500)
			return
		}
		resp.Body.Close()

		fmt.Fprintln(w, "Runner registration triggered.")
	})

	http.HandleFunc("/register-to-runner-api", func(w http.ResponseWriter, r *http.Request) {
		token, err := github.GetRunnerRegistrationToken()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		payload := map[string]string{
			"token": token,
			"url":   fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO")),
		}

		data, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:8081/register-api", "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("❌ Failed to send token to runner: %v", err)
			http.Error(w, "failed to send token", 500)
			return
		}
		defer resp.Body.Close()

		fmt.Fprintln(w, "Runner API registration triggered.")
	})

	http.HandleFunc("/register-hybrid", func(w http.ResponseWriter, r *http.Request) {
		token, err := github.GetRunnerRegistrationToken()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		payload := map[string]string{
			"token": token,
			"url":   fmt.Sprintf("https://github.com/%s/%s", os.Getenv("GITHUB_OWNER"), os.Getenv("GITHUB_REPO")),
		}

		data, _ := json.Marshal(payload)
		resp, err := http.Post("http://localhost:8081/register-hybrid", "application/json", bytes.NewBuffer(data))
		if err != nil {
			http.Error(w, "failed to send token to runner", 500)
			return
		}
		defer resp.Body.Close()

		fmt.Fprintln(w, "Hybrid runner registration triggered.")
	})

	// Daftar routes (semua sebelum ListenAndServe)
	http.HandleFunc("/github/webhook", github.WebhookHandler)
	controller.RegisterHTTPRoutes()   // /jobs
	controller.RegisterRunnerRoutes() // /heartbeat, /runners
	controller.RegisterResultRoute()  // /job/result
	controller.ExposeMetrics()        //metrics
	// controller.AutoResetStuckRunners() // add this line ✅

	// Jalankan dispatcher background
	controller.StartDispatcher()
	// controller.AutoResetStuckRunners()

	port := ":8080"
	log.Printf("🚀 Towerd (Integration + Queue + Dispatcher + Callback) running on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
