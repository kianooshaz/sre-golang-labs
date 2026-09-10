package app

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"sre-labs/19-k8s-pod-manager/api"
	"sre-labs/19-k8s-pod-manager/internal/kube"
	"sre-labs/19-k8s-pod-manager/internal/store"
	"sre-labs/19-k8s-pod-manager/internal/worker"
)

func Run() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// In-memory state shared by the API and the workers.
	st := store.New()

	// Kubernetes client (uses the current context / kubeconfig).
	kc, err := kube.New()
	if err != nil {
		slog.Error("failed to create kubernetes client", "error", err)
		os.Exit(1)
	}

	// Worker pool: 2 workers in 2 goroutines.
	pool := worker.NewPool(st, kc, worker.DefaultConfig())
	pool.Start()
	defer pool.Stop()

	// HTTP API: create pod / pod status.
	go api.StartServer(st, kc)

	// Wait for Ctrl+C or SIGTERM, then stop the workers.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	slog.Info("pod manager is running", "addr", ":8080")
	<-sig
	slog.Info("shutting down")
}
