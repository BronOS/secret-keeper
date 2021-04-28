package apiserver

import (
	"net/http"
	"time"
)

func HandlePing(startedAt time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d := time.Since(startedAt)

		writeJSON(w, &map[string]interface{}{
			"status": "OK",
			"uptime": int64(d.Seconds()),
		}, http.StatusOK)
	}
}
