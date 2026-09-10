package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"sre-labs/19-k8s-pod-manager/internal/kube"
	"sre-labs/19-k8s-pod-manager/internal/store"
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
func (h *handlers) createPod(w http.ResponseWriter, r *http.Request) {
	var req createPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// If the app already tracks this pod, do not create it twice.
	if _, ok := h.store.GetPod(req.Name); ok {
		writeError(w, http.StatusConflict, "pod already tracked by pod-manager")
		return
	}

	if err := h.kube.CreatePod(r.Context(), req.Name, req.Image); err != nil {
		if apierrors.IsAlreadyExists(err) {
			writeError(w, http.StatusConflict, "pod already exists in cluster")
			return
		}
		writeError(w, http.StatusInternalServerError, "create pod: "+err.Error())
		return
	}

	pod := store.Pod{
		Name:      req.Name,
		Image:     req.Image,
		CreatedAt: time.Now().UTC(),
	}
	h.store.AddPod(pod)

	// Kick the worker pool so the first status arrives quickly.
	h.store.Enqueue(req.Name)

	writeJSON(w, http.StatusCreated, pod)
}

// listPods handles GET /api/v1/pods
func (h *handlers) listPods(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.store.ListPods())
}

// podStatus handles GET /api/v1/pods/{name}/status
// It only reads the in-memory status map written by the workers.
func (h *handlers) podStatus(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	if _, ok := h.store.GetPod(name); !ok {
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
	if s[0] == '-' || s[len(s)-1] == '-' {
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
