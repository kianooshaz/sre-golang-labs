package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/codes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"sre-labs/21-observibility/internal/kube"
	"sre-labs/21-observibility/internal/obs"
	"sre-labs/21-observibility/internal/store"
)

type handlers struct {
	store *store.Store
	kube  kube.Client
}

func (h *handlers) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type createPodRequest struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

func (r *createPodRequest) validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if len(r.Name) > 63 || !isDNS1123Label(r.Name) {
		return errors.New("name must be a valid kubernetes name (lowercase alphanumeric and '-', max 63 chars)")
	}
	if r.Image == "" {
		r.Image = defaultImage
	}
	return nil
}

// createPod handles POST /api/v1/pods
// Body: {"name": "nginx-1", "image": "nginx:alpine"}
// It creates the pod in the cluster and adds it to the in-memory pod
// list so the workers start checking it.
//
// Every outcome increments PodCreations with a distinct label
// (created / conflict / error / invalid), so dashboards can show
// success vs rejection vs failure rates at a glance.
func (h *handlers) createPod(w http.ResponseWriter, r *http.Request) {
	ctx, span, log := obs.StartSpan(r.Context(), "handlers.createPod")
	defer func() { obs.EndSpan(span, nil) }()

	var req createPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		obs.PodCreations.WithLabelValues("invalid").Inc()
		span.SetAttributes(obs.AttrsOf("outcome", "invalid")...)
		span.SetStatus(codes.Error, err.Error())
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.validate(); err != nil {
		obs.PodCreations.WithLabelValues("invalid").Inc()
		span.SetAttributes(obs.AttrsOf("outcome", "invalid")...)
		span.SetStatus(codes.Error, err.Error())
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	log = log.With("pod", req.Name, "image", req.Image)
	span.SetAttributes(obs.AttrsOf("pod.name", req.Name, "pod.image", req.Image)...)

	// If the app already tracks this pod, do not create it twice.
	if _, ok := h.store.GetPod(req.Name); ok {
		obs.PodCreations.WithLabelValues("conflict").Inc()
		log.Warn("create rejected: pod already tracked")
		writeError(w, http.StatusConflict, "pod already tracked by pod-manager")
		return
	}

	start := time.Now()
	err := h.kube.CreatePod(ctx, req.Name, req.Image)
	obs.PodCreateDuration.Observe(time.Since(start).Seconds())
	if err != nil {
		obs.PodCreations.WithLabelValues("error").Inc()
		log.Error("create pod in cluster failed", "error", err, "kube_duration_ms", time.Since(start).Milliseconds())
		if apierrors.IsAlreadyExists(err) {
			// EndSpan inside defer would mark success; record the outcome here.
			span.SetAttributes(obs.AttrsOf("outcome", "conflict")...)
			writeError(w, http.StatusConflict, "pod already exists in cluster")
			return
		}
		span.SetAttributes(obs.AttrsOf("outcome", "error")...)
		writeError(w, http.StatusInternalServerError, "create pod: "+err.Error())
		return
	}

	pod := store.Pod{
		Name:      req.Name,
		Image:     req.Image,
		CreatedAt: time.Now().UTC(),
	}
	// Remember the creating trace so worker status-checks can link to it.
	h.store.AddPod(pod, span.SpanContext())
	obs.TrackedPods.Set(float64(len(h.store.ListPods())))

	// Kick the worker pool so the first status arrives quickly.
	h.store.Enqueue(req.Name)

	obs.PodCreations.WithLabelValues("created").Inc()
	span.SetAttributes(obs.AttrsOf("outcome", "created")...)
	log.Info("pod created", "kube_duration_ms", time.Since(start).Milliseconds())
	writeJSON(w, http.StatusCreated, pod)
}

// listPods handles GET /api/v1/pods
func (h *handlers) listPods(w http.ResponseWriter, r *http.Request) {
	_, span, log := obs.StartSpan(r.Context(), "handlers.listPods")
	defer obs.EndSpan(span, nil)
	pods := h.store.ListPods()
	log.Debug("listed pods", "count", len(pods))
	writeJSON(w, http.StatusOK, pods)
}

// podStatus handles GET /api/v1/pods/{name}/status
// It only reads the in-memory status map written by the workers.
func (h *handlers) podStatus(w http.ResponseWriter, r *http.Request) {
	_, span, log := obs.StartSpan(r.Context(), "handlers.podStatus")
	defer obs.EndSpan(span, nil)
	name := r.PathValue("name")
	span.SetAttributes(obs.AttrsOf("pod.name", name)...)

	if _, ok := h.store.GetPod(name); !ok {
		log.Warn("status requested for untracked pod", "pod", name)
		writeError(w, http.StatusNotFound, "pod is not tracked by pod-manager")
		return
	}

	st, ok := h.store.GetStatus(name)
	if !ok {
		// Created but not checked by a worker yet.
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "pending", "detail": "worker has not checked this pod yet"})
		return
	}

	writeJSON(w, http.StatusOK, st)
}

// isDNS1123Label implements the kubernetes name rule for labels:
// lowercase alphanumeric or '-', start/end alphanumeric, max 63 chars.
func isDNS1123Label(s string) bool {
	if s == "" || s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-':
		default:
			return false
		}
	}
	return true
}
