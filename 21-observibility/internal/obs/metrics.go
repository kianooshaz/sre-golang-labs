package obs

import (
	"context"
	"expvar"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics are the application-level signals Prometheus scrapes from /metrics.
// All of them use the same label dimensions (method, route, outcome, pod)
// so dashboards can slice the data the same way everywhere.
var (
	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "podmanager",
			Name:      "http_requests_total",
			Help:      "Total HTTP requests handled, by method, route and status code.",
		},
		[]string{"method", "route", "code"},
	)

	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "podmanager",
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency in seconds, by method and route.",
			// 1ms .. 10s spread, good enough for API latencies.
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route"},
	)

	PodCreations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "podmanager",
			Name:      "pod_creations_total",
			Help:      "Pod create attempts, by outcome (created / conflict / error / invalid).",
		},
		[]string{"outcome"},
	)

	PodCreateDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "podmanager",
			Name:      "pod_create_duration_seconds",
			Help:      "Time spent creating a pod in Kubernetes (API round trip).",
			Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
	)

	PodChecks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "podmanager",
			Name:      "pod_status_checks_total",
			Help:      "Pod status checks performed by the worker pool, by outcome.",
		},
		[]string{"worker", "outcome"},
	)

	PodCheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "podmanager",
			Name:      "pod_status_check_duration_seconds",
			Help:      "Time a worker spends on one pod status check.",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"worker"},
	)

	QueueLength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "podmanager",
			Name:      "workqueue_length",
			Help:      "Number of pod names waiting in the worker workqueue.",
		},
	)

	TrackedPods = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "podmanager",
			Name:      "tracked_pods",
			Help:      "Number of pods currently tracked by pod-manager.",
		},
	)

	QueueDropped = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "podmanager",
			Name:      "workqueue_dropped_total",
			Help:      "Enqueue attempts dropped because the workqueue was full.",
		},
	)
)

// registry holds only our collectors plus the go/runtime ones we care about;
// a dedicated registry avoids the default registry's noisy process metrics.
var registry = prometheus.NewRegistry()

func initMetrics(_ Config) {
	registry.MustRegister(
		HTTPRequests,
		HTTPDuration,
		PodCreations,
		PodCreateDuration,
		PodChecks,
		PodCheckDuration,
		QueueLength,
		TrackedPods,
		QueueDropped,
		// Go runtime gauges (goroutines, GC, memory) — great for SRE dashboards.
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
}

// MetricsHandler serves the Prometheus scrape endpoint.
func MetricsHandler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		registry,
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
}

// RuntimeSampler publishes goroutine and heap gauges to /debug/vars
// (expvar) every tick until ctx is cancelled. A zero-dependency way to
// peek at runtime health while debugging locally.
func RuntimeSampler(ctx context.Context, tick time.Duration) {
	goroutines := expvar.NewInt("goroutines")
	heapAlloc := expvar.NewInt("heap_alloc_bytes")

	go func() {
		t := time.NewTicker(tick)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				var m runtime.MemStats
				goroutines.Set(int64(runtime.NumGoroutine()))
				runtime.ReadMemStats(&m)
				heapAlloc.Set(int64(m.HeapAlloc))
			}
		}
	}()
}
