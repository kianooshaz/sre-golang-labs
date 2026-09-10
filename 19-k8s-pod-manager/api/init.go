package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"sre-labs/19-k8s-pod-manager/internal/kube"
	"sre-labs/19-k8s-pod-manager/internal/store"
)

const addr = ":8080"

// StartServer runs the HTTP API (blocking).
func StartServer(st *store.Store, kc kube.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := kc.EnsureNamespace(ctx); err != nil {
		slog.Warn("could not ensure namespace", "error", err)
	}
	cancel()

	e := newRouter(st, kc)
	slog.Info("starting http server", "addr", addr)
	if err := http.ListenAndServe(addr, e); err != nil {
		slog.Error("http server failed", "error", err)
	}
}
