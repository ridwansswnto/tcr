package main

import (
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

	// Daftar routes (semua sebelum ListenAndServe)
	http.HandleFunc("/github/webhook", github.WebhookHandler)
	controller.RegisterHTTPRoutes()    // /jobs
	controller.RegisterRunnerRoutes()  // /heartbeat, /runners
	controller.RegisterResultRoute()   // /job/result
	controller.ExposeMetrics()         //metrics
	controller.AutoResetStuckRunners() // add this line ✅

	// Jalankan dispatcher background
	controller.StartDispatcher()
	// controller.AutoResetStuckRunners()

	port := ":8080"
	log.Printf("🚀 Towerd (Integration + Queue + Dispatcher + Callback) running on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
