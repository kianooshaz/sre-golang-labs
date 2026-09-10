// Package obs wires the three observability pillars for the app:
//
//   - logs:   rich slog JSON logs with request_id / trace_id injected
//   - traces: OpenTelemetry spans (exportable to OTLP / Tempo / Jaeger)
//   - metrics: Prometheus counters, histograms and gauges on /metrics
//
// Every component is initialized once from Init and shut down from Shutdown.
package obs

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// Config controls how the observability stack is set up.
type Config struct {
	ServiceName    string // used as the service.name resource attribute
	ServiceVersion string // used as the service.version resource attribute
	Env            string // dev / staging / prod — added to logs and spans
	OTLPEndpoint   string // host:port of an OTLP gRPC collector; empty disables tracing export
	LogLevel       string // debug / info / warn / error
	ExtraLogAttrs  []any  // permanent attrs appended to every log line
}

// state holds the live global components; Init populates it.
var state struct {
	logger      *slog.Logger
	shutdownFns []func(context.Context) error
}

// Init installs the structured logger and the trace provider.
// Call it once at startup; pair it with Shutdown on the way out.
func Init(ctx context.Context, cfg Config) error {
	state.logger = newLogger(cfg)
	initMetrics(cfg)
	if err := initTracing(ctx, cfg); err != nil {
		return err
	}

	state.logger.Info("observability initialized",
		"otlp_endpoint", cfg.OTLPEndpoint,
		"tracing", strings.TrimSpace(cfg.OTLPEndpoint) != "",
	)
	return nil
}

// Shutdown flushes traces (and any other async exporters) and blocks
// until the exporters finish or ctx expires.
func Shutdown(ctx context.Context) error {
	var firstErr error
	for _, fn := range state.shutdownFns {
		if err := fn(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// Logger returns the configured application logger.
func Logger() *slog.Logger {
	if state.logger == nil {
		return slog.Default()
	}
	return state.logger
}

// newLogger builds a JSON slog logger with env/service detail on every line.
func newLogger(cfg Config) *slog.Logger {
	lvl := slog.LevelInfo
	switch strings.ToLower(cfg.LogLevel) {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	}

	// ReplaceAttr is the hook that enriches every record before it is written:
	// level names become uppercase and the time is formatted the same everywhere.
	opts := &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Value.Kind() {
			case slog.KindDuration:
				// Durations print as "12.3ms" instead of a huge nanosecond integer.
				a.Value = slog.StringValue(a.Value.Duration().String())
			case slog.KindTime:
				a.Value = slog.StringValue(a.Value.Time().UTC().Format("2006-01-02T15:04:05.000Z"))
			}
			return a
		},
	}

	h := slog.NewJSONHandler(os.Stdout, opts)

	// Base attributes every log line carries: service identity, env, pid.
	base := []any{
		slog.String("service", cfg.ServiceName),
		slog.String("env", cfg.Env),
		slog.String("version", cfg.ServiceVersion),
		slog.Int("pid", os.Getpid()),
	}
	base = append(base, cfg.ExtraLogAttrs...)

	return slog.New(h).With(base...)
}
