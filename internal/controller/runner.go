package controller

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Runner struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	Port     string    `json:"port"`
	LastSeen time.Time `json:"last_seen"`
	IsBusy   bool      `json:"is_busy"`
}

var (
	runners   = make(map[string]*Runner)
	runnersMu sync.Mutex
)

// HeartbeatHandler menerima ping dari runner
func HeartbeatHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	port := r.URL.Query().Get("port")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	if port == "" {
		port = "8081" // default runner port
	}

	// ambil hanya IP dari RemoteAddr (tanpa port ephemeral)
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Printf("⚠️ Cannot parse host from RemoteAddr: %v", err)
		host = "localhost"
	}

	runnersMu.Lock()
	defer runnersMu.Unlock()

	if runner, exists := runners[id]; exists {
		runner.LastSeen = time.Now()
		runner.Port = port
	} else {
		runners[id] = &Runner{
			ID:       id,
			Address:  host,
			Port:     port,
			LastSeen: time.Now(),
		}
		log.Printf("🟢 Runner registered: %s (%s:%s)", id, host, port)
	}

	w.WriteHeader(http.StatusOK)
}

// GetIdleRunner mencari runner yang belum sibuk
func GetIdleRunner() *Runner {
	runnersMu.Lock()
	defer runnersMu.Unlock()

	for _, r := range runners {
		if !r.IsBusy && time.Since(r.LastSeen) < 30*time.Second {
			return r
		}
	}
	return nil
}

// MarkRunnerBusy menandai runner sedang sibuk

func MarkRunnerBusy(id string, busy bool) {
	runnersMu.Lock()
	defer runnersMu.Unlock()

	if r, ok := runners[id]; ok {
		r.IsBusy = busy
		if !busy {
			log.Printf("🟢 Runner %s is now idle", id)
		} else {
			log.Printf("🔴 Runner %s marked busy", id)
		}
	}
}

// RegisterRunnerRoutes untuk endpoint /heartbeat
func RegisterRunnerRoutes() {
	http.HandleFunc("/heartbeat", HeartbeatHandler)
	http.HandleFunc("/runners", func(w http.ResponseWriter, r *http.Request) {
		runnersMu.Lock()
		defer runnersMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(runners)
	})
}
