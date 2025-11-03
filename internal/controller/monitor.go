package controller

import (
	"log"
	"time"
)

// AutoResetStuckRunners berjalan di background dan akan otomatis
// menandai runner sebagai idle jika sudah busy tapi tidak ada job aktif.
func AutoResetStuckRunners() {
	go func() {
		for {
			time.Sleep(30 * time.Second) // interval pengecekan setiap 30 detik

			runnersMu.Lock()
			for _, r := range runners {
				if r.IsBusy {
					active := false

					// Cek apakah masih ada job aktif (dispatched/running)
					jobQueueMu.Lock()
					for _, j := range jobQueue {
						if j.Status == "dispatched" || j.Status == "running" {
							active = true
							break
						}
					}
					jobQueueMu.Unlock()

					// Kalau runner busy tapi gak ada job aktif → reset jadi idle
					if !active {
						r.IsBusy = false
						log.Printf("🧹 Auto-reset runner %s (no active jobs)", r.ID)
					}
				}
			}
			runnersMu.Unlock()
		}
	}()
}
