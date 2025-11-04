package core

import "log"

// Channel untuk komunikasi antar package tanpa import langsung
var JobQueue = make(chan Job, 100)

// AddJob = enqueue job ke global channel
func AddJob(j Job) {
	select {
	case JobQueue <- j:
		log.Printf("📥 Enqueued job: %s (%s/%s)", j.JobName, j.RepoOwner, j.RepoName)
	default:
		log.Printf("⚠️ Queue full, dropping job: %s", j.JobName)
	}
}
