package api

import (
	"encoding/json"
	"net/http"

	"sre-labs/19-k8s-pod-manager/internal/kube"
	"sre-labs/19-k8s-pod-manager/internal/store"
)

// defaultImage is used when the request has no image.
const defaultImage = "nginx:alpine"

func newRouter(st *store.Store, kc kube.Client) http.Handler {
	mux := http.NewServeMux()
	h := &handlers{store: st, kube: kc}

	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /api/v1/pods", h.createPod)
	mux.HandleFunc("GET /api/v1/pods", h.listPods)
	mux.HandleFunc("GET /api/v1/pods/{name}/status", h.podStatus)

	return logRequests(mux)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
