package github

import (
	"encoding/json"
	"net/http"
)

// Handler untuk /github/token (dipanggil agent)
func TokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := GetRunnerRegistrationToken()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"token": token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
