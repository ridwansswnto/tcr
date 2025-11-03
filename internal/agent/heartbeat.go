package agent

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

func (a *Agent) HeartbeatLoop() {
	for {
		data := map[string]interface{}{
			"instance":  a.config.InstanceName,
			"runners":   len(a.runners),
			"timestamp": time.Now(),
		}
		b, _ := json.Marshal(data)
		http.Post(a.config.TowerURL+"/vm/heartbeat", "application/json", bytes.NewBuffer(b))
		time.Sleep(time.Duration(a.config.HeartbeatInterval) * time.Second)
	}
}
