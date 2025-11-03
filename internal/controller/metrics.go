package controller

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	JobTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tcr_jobs_total",
			Help: "Total jobs by status",
		},
		[]string{"status"},
	)
	JobsInQueue = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "tcr_jobs_in_queue",
			Help: "Number of jobs currently in queue",
		},
	)
	JobDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "tcr_job_duration_seconds",
			Help:    "Job duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.5, 2, 10),
		},
	)
	RunnersTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "tcr_runners_total",
			Help: "Number of registered runners",
		},
	)
	RunnersIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "tcr_runners_idle",
			Help: "Number of idle runners",
		},
	)
	DispatchErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "tcr_dispatch_errors_total",
			Help: "Number of failed dispatch attempts",
		},
	)
)

func init() {
	prometheus.MustRegister(JobTotal, JobsInQueue, JobDuration, RunnersTotal, RunnersIdle, DispatchErrors)
}

// ExposeMetrics registers /metrics endpoint on the default mux (or explicit one)
func ExposeMetrics() {
	http.Handle("/metrics", promhttp.Handler())
	// optionally run separate listener if you want different port for metrics
	// go http.ListenAndServe(addr, nil)
}

// Helper functions to update metrics — call from job queue / dispatcher / runners
func incJobStatus(status string) {
	JobTotal.WithLabelValues(status).Inc()
	updateJobsInQueue()
}

func updateJobsInQueue() {
	jobQueueMu.Lock()
	JobsInQueue.Set(float64(len(jobQueue)))
	jobQueueMu.Unlock()
}

func observeJobDuration(d time.Duration) {
	JobDuration.Observe(d.Seconds())
}

func updateRunnerGauges() {
	runnersMu.Lock()
	total := float64(len(runners))
	idleCount := 0
	for _, r := range runners {
		if !r.IsBusy && time.Since(r.LastSeen) < 30*time.Second {
			idleCount++
		}
	}
	runnersMu.Unlock()
	RunnersTotal.Set(total)
	RunnersIdle.Set(float64(idleCount))
}
