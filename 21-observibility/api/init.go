package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"sre-labs/21-observibility/internal/kube"
	"sre-labs/21-observibility/internal/obs"
	"sre-labs/21-observibility/internal/store"
)

const addr = ":8080"

// StartServer runs the HTTP API. It returns when ctx is cancelled and the
// in-flight requests have drained (graceful shutdown), or on a fatal error.
func StartServer(ctx context.Context, st *store.Store, kc kube.Client) error {
	ensureCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	if err := kc.EnsureNamespace(ensureCtx); err != nil {
		obs.Logger().Warn("could not ensure namespace", "error", err)
	}
	cancel()

	srv := &http.Server{
		Addr:              addr,
		Handler:           newRouter(st, kc),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Shutdown when the app context is cancelled (SIGINT/SIGTERM in app.Run).
	errCh := make(chan error, 1)
	go func() {
		obs.Logger().Info("starting http server", "addr", addr,
			"endpoints", []string{"/healthz", "/api/v1/pods", "/metrics", "/debug/vars"})
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		obs.Logger().Info("draining http connections", "addr", addr)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
