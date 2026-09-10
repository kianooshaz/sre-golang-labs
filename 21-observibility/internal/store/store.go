// Package store keeps the in-memory state of the app:
// the list of managed pods and the latest status fetched by the workers.
// The API writes to it, the workers read from it and write statuses back,
// so every access is guarded by a RWMutex.
package store

import (
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"

	"sre-labs/21-observibility/internal/obs"
)

// Pod is one entry of the in-memory pod list.
type Pod struct {
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	CreatedAt time.Time `json:"created_at"`
}

// Status is the latest pod status fetched from Kubernetes by a worker.
type Status struct {
	PodName   string    `json:"pod_name"`
	Phase     string    `json:"phase"`
	Ready     string    `json:"ready"` // e.g. "1/1"
	Restarts  int32     `json:"restarts"`
	Node      string    `json:"node,omitempty"`
	IP        string    `json:"ip,omitempty"`
	FetchedAt time.Time `json:"fetched_at"`
	Error     string    `json:"error,omitempty"`
}

// Store is the shared in-memory state:
//   - pods:     the list of pods we manage
//   - statuses: latest status per pod, written by the workers
//   - queue:    pod names waiting to be checked by a worker
//   - origins:  trace context captured at pod creation, so the workers can
//     link their status-check spans back to the creating trace
type Store struct {
	mu       sync.RWMutex
	pods     map[string]Pod
	statuses map[string]Status
	origins  map[string]trace.SpanContext
	queue    chan string
	stopCh   chan struct{}
	stopOnce sync.Once
}

func New() *Store {
	return &Store{
		pods:     make(map[string]Pod),
		statuses: make(map[string]Status),
		origins:  make(map[string]trace.SpanContext),
		queue:    make(chan string, 128),
		stopCh:   make(chan struct{}),
	}
}

// AddPod registers a pod in the in-memory list.
// It returns false if the pod is already tracked.
func (s *Store) AddPod(p Pod, origin trace.SpanContext) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.pods[p.Name]; ok {
		return false
	}
	s.pods[p.Name] = p
	s.origins[p.Name] = origin

	return true
}

// GetPod returns a pod from the in-memory list.
func (s *Store) GetPod(name string) (Pod, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.pods[name]
	return p, ok
}

// ListPods returns a snapshot of the in-memory pod list.
func (s *Store) ListPods() []Pod {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Pod, 0, len(s.pods))
	for _, p := range s.pods {
		out = append(out, p)
	}
	return out
}

// SetStatus stores the latest status of a pod in the in-memory map.
func (s *Store) SetStatus(st Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statuses[st.PodName] = st
}

// GetStatus returns the latest status recorded for a pod.
func (s *Store) GetStatus(name string) (Status, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.statuses[name]
	return st, ok
}

// Origin returns the trace context captured when the pod was created, so a
// worker can link its status-check span back to the creating trace.
func (s *Store) Origin(name string) (trace.SpanContext, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.origins[name]
	return sc, ok
}

// Enqueue queues a pod name for the next free worker (never blocks).
// Dropped enqueues (full queue) are counted in workqueue_dropped_total.
func (s *Store) Enqueue(name string) {
	select {
	case s.queue <- name:
		obs.QueueLength.Set(float64(len(s.queue)))
	default:
		obs.QueueDropped.Inc()
	}
}

// Queue is the read side of the work queue, consumed by the workers.
func (s *Store) Queue() <-chan string { return s.queue }

// Stop signals the workers to exit; it is safe to call more than once.
func (s *Store) Stop() {
	s.stopOnce.Do(func() { close(s.stopCh) })
}

// Done is closed when the workers should stop.
func (s *Store) Done() <-chan struct{} { return s.stopCh }
