// Package obs is the observability bonus (lab 21): Prometheus metrics for
// the ingest pipeline. Optional — finish the core first.
package obs

import (
	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics owns logan's counters. Names are fixed so a dashboard can be built
// against them:
//
//	logan_lines_ingested_total   counter — lines accepted into the queue
//	logan_lines_rejected_total   counter — lines refused (429 backpressure)
//	logan_lines_malformed_total  counter — corrupt lines counted by the pool
//	logan_queue_length           gauge   — current ingest queue depth
type Metrics struct {
	Ingested  prometheus.Counter
	Rejected  prometheus.Counter
	Malformed prometheus.Counter
	QueueLen  prometheus.Gauge
}

// New declares the four collectors and registers them with
// prometheus.DefaultRegisterer.
// TODO(student): prometheus.NewCounter(prometheus.CounterOpts{Name: ..., Help: ...})
// and friends, then prometheus.MustRegister each.
func New() *Metrics {
	panic("TODO(student): declare and register the collectors")
}

// Handler returns the handler for GET /metrics.
func Handler() echo.HandlerFunc {
	return echo.WrapHandler(promhttp.Handler())
}
