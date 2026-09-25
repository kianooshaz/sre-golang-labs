// Package server exposes logan's HTTP API with echo (lab 11): an ingest
// firehose with backpressure, three query endpoints, and an optional
// Prometheus /metrics (bonus, lab 21).
package server

import (
	"context"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"logan/internal/alert"
	"logan/internal/config"
	"logan/internal/obs"
	"logan/internal/pool"
	"logan/internal/store"
)

// Server owns the echo instance and logan's collaborators.
type Server struct {
	e    *echo.Echo
	cfg  config.Config
	pool *pool.Pool
	st   store.Store
	ev   alert.Evaluator
	m    *obs.Metrics // nil unless you do the observability bonus
}

// New wires middleware and routes. It is implemented on purpose — read it to
// see the object graph; the handler bodies are the student's work.
func New(cfg config.Config, pl *pool.Pool, st store.Store, ev alert.Evaluator, m *obs.Metrics) *Server {
	e := echo.New()
	e.Use(middleware.Recover())

	s := &Server{e: e, cfg: cfg, pool: pl, st: st, ev: ev, m: m}
	s.routes()
	return s
}

// routes registers the API. Paths and verbs are fixed — the README contract
// and the sample clients rely on them.
func (s *Server) routes() {
	s.e.POST("/ingest", s.handleIngest)
	s.e.GET("/stats", s.handleStats)
	s.e.GET("/patterns", s.handlePatterns)
	s.e.GET("/alerts", s.handleAlerts)
	if s.m != nil {
		s.e.GET("/metrics", obs.Handler())
	}
}

// handleIngest answers POST /ingest. The body is newline-delimited raw log
// lines (either format, corrupt lines allowed). For each line: Submit to the
// pool. Contract:
//
//   - every line accepted (or body empty): 202
//     {"status":"queued","lines":N,"rejected":0}
//   - some accepted, some refused by pool.ErrQueueFull: still 202 with both
//     counts — partial acceptance beats losing the batch
//   - every line refused: 429 {"error":"queue_full","rejected":N}
//
// TODO(student): split the body on '\n', skip empty lines, Submit each, and
// translate ErrQueueFull into the counts above.
func (s *Server) handleIngest(c *echo.Context) error {
	panic("TODO(student): POST /ingest — the backpressure story")
}

// handleStats answers GET /stats from store.Stats:
//
//	{"services":[{"name":"billing","total":1200,"errors":33,"error_rate":0.0275}]}
//
// TODO(student)
func (s *Server) handleStats(c *echo.Context) error {
	panic("TODO(student): GET /stats")
}

// handlePatterns answers GET /patterns?top=10&service=billing from
// store.TopPatterns:
//
//	[{"signature":"payment failed for order <NUM>","occurrences":57,"last_seen":"..."}]
//
// TODO(student): bind `top` (default 10, cap at 100) and optional `service`.
func (s *Server) handlePatterns(c *echo.Context) error {
	panic("TODO(student): GET /patterns?top=N")
}

// handleAlerts answers GET /alerts?open=true from store.Alerts:
//
//	[{"id":1,"service_id":2,"pattern_id":null,"rate":0.42,"threshold":0.3,
//	  "window_seconds":300,"created_at":"...","resolved_at":null}]
//
// TODO(student)
func (s *Server) handleAlerts(c *echo.Context) error {
	panic("TODO(student): GET /alerts")
}

// Start blocks until ctx is canceled (Ctrl+C/SIGTERM — echo then drains
// in-flight requests for up to ShutdownGrace) or the listener fails.
// In echo v5 shutdown is context-driven: there is no separate Shutdown
// method; whatever must outlive the HTTP server (the pool drain) runs in the
// caller after Start returns.
func (s *Server) Start(ctx context.Context) error {
	sc := echo.StartConfig{
		Address:         s.cfg.Server.Addr,
		HideBanner:      true,
		GracefulTimeout: s.cfg.Server.ShutdownGrace,
	}
	return sc.Start(ctx, s.e)
}
