package telemetry

import (
	"log"
	"time"
)

// LogGraph writes a compact graph execution timing log.
func LogGraph(name string, start time.Time, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	log.Printf("[graph] name=%s duration_ms=%d error=%q", name, time.Since(start).Milliseconds(), msg)
}

// LogNode writes a compact node execution timing log.
func LogNode(name string, start time.Time, err error) {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	log.Printf("[node] name=%s duration_ms=%d error=%q", name, time.Since(start).Milliseconds(), msg)
}
