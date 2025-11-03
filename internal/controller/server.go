package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type Agent struct {
	ID       string
	Address  string
	LastSeen time.Time
	IsActive bool
}

var (
	agents = make(map[string]*Agent)
	mu     sync.Mutex
)

func StartServer(addr string) {
	http.HandleFunc("/heartbeat", handleHeartbeat)
	http.HandleFunc("/agents", handleAgents)

	go monitorAgents()

	log.Printf("Towerd listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	ip := r.RemoteAddr
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if a, ok := agents[id]; ok {
		a.LastSeen = time.Now()
		a.IsActive = true
	} else {
		agents[id] = &Agent{ID: id, Address: ip, LastSeen: time.Now(), IsActive: true}
		log.Printf("🆕 Registered new agent: %s (%s)", id, ip)
	}

	w.WriteHeader(http.StatusOK)
}

func handleAgents(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	json.NewEncoder(w).Encode(agents)
}

func monitorAgents() {
	for {
		time.Sleep(10 * time.Second)
		now := time.Now()

		mu.Lock()
		for id, a := range agents {
			if a.IsActive && now.Sub(a.LastSeen) > time.Minute {
				log.Printf("⚠️ Agent %s idle > 1m, sending shutdown", id)
				go sendShutdown(a)
				a.IsActive = false
			}
		}
		mu.Unlock()
	}
}

func sendShutdown(a *Agent) {
	url := "http://" + a.Address + "/shutdown"
	resp, err := http.Post(url, "application/json", nil)
	if err != nil {
		log.Printf("❌ Failed to shutdown %s: %v", a.ID, err)
		return
	}
	defer resp.Body.Close()
	log.Printf("✅ Sent shutdown to %s", a.ID)
}
