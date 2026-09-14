// Package worker implements the worker pool that polls pod statuses.
//
// The pool runs N workers (default 2), each in its own goroutine.
// Workers consume pod names from the shared queue in the store and,
// on every tick of the ticker (default every 2 minutes), the whole
// in-memory pod list is queued again.
package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"sre-labs/19-k8s-pod-manager/internal/kube"
	"sre-labs/19-k8s-pod-manager/internal/store"
)

type Config struct {
	Workers  int           // number of goroutines in the pool
	Interval time.Duration // how often every pod is checked
}

func DefaultConfig() Config {
	return Config{Workers: 2, Interval: 2 * time.Minute}
}

type Pool struct {
	store  *store.Store
	kube   kube.Client
	cfg    Config
	ticker *time.Ticker
}

func NewPool(st *store.Store, kc kube.Client, cfg Config) *Pool {
	return &Pool{store: st, kube: kc, cfg: cfg, ticker: time.NewTicker(cfg.Interval)}
}

// Start launches the worker goroutines and then blocks, re-queueing the whole
// in-memory pod list on every tick, until the store is stopped.
// Run it in its own goroutine: `go pool.Start()`.
func (p *Pool) Start() {
	for i := 1; i <= p.cfg.Workers; i++ {
		go p.worker(i)
	}

	slog.Info("worker pool started", "workers", p.cfg.Workers, "interval", p.cfg.Interval)

	for {
		select {
		case <-p.store.Done():
			return
		case <-p.ticker.C:
			if n := p.store.EnqueueAll(); n > 0 {
				slog.Info("queued all pods for status check", "count", n)
			}
		}
	}
}

// Stop stops the ticker and signals the workers to exit.
func (p *Pool) Stop() {
	p.ticker.Stop()
	p.store.Stop()
}

// worker is the body of one pool goroutine.
func (p *Pool) worker(id int) {
	log := slog.Default().With("worker", id)
	log.Info("worker started")
	defer log.Info("worker stopped")

	for {
		select {
		case <-p.store.Done():
			return

		case name := <-p.store.Queue():
			p.checkPod(log, name)
		}
	}
}

// checkPod fetches the pod status from Kubernetes and writes it
// into the in-memory status map.
func (p *Pool) checkPod(log *slog.Logger, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pod, err := p.kube.GetPod(ctx, name)
	st := store.Status{PodName: name, FetchedAt: time.Now().UTC()}

	switch {
	case err == nil:
		st.Phase = string(pod.Status.Phase)
		st.Ready = readyCount(pod)
		st.Restarts = totalRestarts(pod)
		st.Node = pod.Spec.NodeName
		st.IP = pod.Status.PodIP
	case apierrors.IsNotFound(err):
		st.Error = "pod not found in cluster"
	default:
		st.Error = err.Error()
	}

	p.store.SetStatus(st)
	log.Info("checked pod", "pod", name, "phase", st.Phase, "ready", st.Ready, "error", st.Error)
}

func readyCount(pod *corev1.Pod) string {
	var ready, total int32
	for _, cs := range pod.Status.ContainerStatuses {
		total++
		if cs.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, total)
}

func totalRestarts(pod *corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}
