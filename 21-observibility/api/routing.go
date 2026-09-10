package api

import (
	"encoding/json"
	"expvar"
	"net/http"

	"sre-labs/21-observibility/internal/kube"
	"sre-labs/21-observibility/internal/obs"
	"sre-labs/21-observibility/internal/store"
)

// defaultImage is used when the request has no image.
const defaultImage = "nginx:alpine"

// patternKey and routeHolder live in middleware.go — this file only fills
// the holder in via withRoute.

func newRouter(st *store.Store, kc kube.Client) http.Handler {
	mux := http.NewServeMux()
	h := &handlers{store: st, kube: kc}

	// Each route is registered with its pattern so the observability
	// middleware can label metrics and spans with the low-cardinality
	// route (e.g. "POST /api/v1/pods") instead of the raw path.
	mux.HandleFunc("GET /healthz", withRoute("GET /healthz", h.health))
	mux.HandleFunc("POST /api/v1/pods", withRoute("POST /api/v1/pods", h.createPod))
	mux.HandleFunc("GET /api/v1/pods", withRoute("GET /api/v1/pods", h.listPods))
	mux.HandleFunc("GET /api/v1/pods/{name}/status", withRoute("GET /api/v1/pods/{name}/status", h.podStatus))

	// Prometheus scrape endpoint — deliberately outside the request
	// middleware so scrape traffic does not pollute the API's own metrics.
	metricsMux := http.NewServeMux()
	metricsMux.Handle("GET /metrics", obs.MetricsHandler())

	// /debug/vars (expvar) for quick runtime health peeks.
	debugMux := http.NewServeMux()
	debugMux.Handle("GET /debug/vars", expvar.Handler())

	return &combinedHandler{
		api:     obsMiddleware(mux),
		metrics: metricsMux,
		debug:   debugMux,
	}
}

// combinedHandler serves /metrics and /debug/vars directly and sends
// everything else through the observed API mux.
type combinedHandler struct {
	api     http.Handler
	metrics http.Handler
	debug   http.Handler
}

func (c *combinedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/metrics":
		c.metrics.ServeHTTP(w, r)
	case "/debug/vars":
		c.debug.ServeHTTP(w, r)
	default:
		c.api.ServeHTTP(w, r)
	}
}

// withRoute records the route pattern in the shared routeHolder the obs
// middleware put in the context, so the middleware can label metrics and
// spans with the low-cardinality route after the handler returns.
func withRoute(pattern string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h, ok := r.Context().Value(patternKey{}).(*routeHolder); ok {
			h.route = pattern
		}
		next.ServeHTTP(w, r)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
