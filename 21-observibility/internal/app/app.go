// Package app wires the whole application together:
// observability, Kubernetes client, worker pool and HTTP API.
package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sre-labs/21-observibility/api"
	"sre-labs/21-observibility/internal/kube"
	"sre-labs/21-observibility/internal/obs"
	"sre-labs/21-observibility/internal/store"
	"sre-labs/21-observibility/internal/worker"
)

// Run boots the app and blocks until SIGINT/SIGTERM, then drains everything
// in reverse order: HTTP server -> worker pool -> trace exporter flush.
func Run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- 1. Observability -----------------------------------------------
	// OTLP_ENDPOINT (host:port) sends spans to a local collector, e.g.:
	//   OTLP_ENDPOINT=localhost:4317 go run .
	// Without it, logs still carry trace_id but spans are not exported.
	cfg := obs.Config{
		ServiceName:    "pod-manager",
		ServiceVersion: "1.0.0",
		Env:            envOr("APP_ENV", "dev"),
		OTLPEndpoint:   os.Getenv("OTLP_ENDPOINT"),
		LogLevel:       envOr("LOG_LEVEL", "info"),
	}
	if err := obs.Init(ctx, cfg); err != nil {
		// Fall back to the default logger: obs is not up yet.
		obs.Logger().Error("failed to init observability", "error", err)
		os.Exit(1)
	}
	log := obs.Logger()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := obs.Shutdown(shutdownCtx); err != nil {
			log.Warn("observability shutdown", "error", err)
		}
	}()

	// Lightweight runtime gauges on /debug/vars.
	obs.RuntimeSampler(ctx, 15*time.Second)

	// ---- 2. In-memory state --------------------------------------------
	st := store.New()

	// ---- 3. Kubernetes client -------------------------------------------
	kc, err := kube.New()
	if err != nil {
		log.Error("failed to create kubernetes client", "error", err)
		os.Exit(1)
	}

	// ---- 4. Worker pool --------------------------------------------------
	pool := worker.NewPool(st, kc, worker.DefaultConfig())
	pool.Start()
	defer pool.Stop()

	// ---- 5. HTTP API ------------------------------------------------------
	apiErr := make(chan error, 1)
	go func() {
		apiErr <- api.StartServer(ctx, st, kc)
	}()

	log.Info("pod manager is running",
		"addr", ":8080",
		"env", cfg.Env,
		"otlp_endpoint", cfg.OTLPEndpoint,
	)

	// Block until signal or fatal API error.
	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-apiErr:
		if err != nil {
			log.Error("http server failed", "error", err)
		}
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
