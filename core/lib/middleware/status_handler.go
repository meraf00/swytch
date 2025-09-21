package middleware

import (
	"net/http"
	"time"

	"github.com/meraf00/swytch/core/lib/logger"
	"github.com/meraf00/swytch/core/lib/respond"
)

type StatusResponse struct {
	StartedAt time.Time `json:"started_at"`
	Uptime    float64   `json:"uptime"`
}

func StatusHandler(startedAt time.Time, log logger.Log) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errorType := r.URL.Query().Get("error")

		switch errorType {
		case "bad-request":
		case "not-found":
		case "validation":
		}

		response := StatusResponse{
			StartedAt: startedAt,
			Uptime:    time.Since(startedAt).Seconds(),
		}

		err := respond.SuccessWithData(w, http.StatusOK, response)
		if err != nil {
			log.Errorf("status_handler_error: %v", err)
		}
	}
}
