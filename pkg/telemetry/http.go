// Package telemetry provides lightweight logging and timing helpers.
package telemetry

import (
	"log"
	"net/http"
	"time"
)

// HTTPEvent captures a small structured log for an HTTP handler.
type HTTPEvent struct {
	Name       string
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
	Error      string
}

// Start returns a timestamp used for duration calculation.
func Start() time.Time { return time.Now() }

// LogHTTP writes a compact structured log line.
func LogHTTP(r *http.Request, status int, start time.Time, name string, err error) {
	e := HTTPEvent{
		Name:       name,
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: status,
		Duration:   time.Since(start),
	}
	if err != nil {
		e.Error = err.Error()
	}
	log.Printf("[http] handler=%s method=%s path=%s status=%d duration_ms=%d error=%q",
		e.Name, e.Method, e.Path, e.StatusCode, e.Duration.Milliseconds(), e.Error)
}
