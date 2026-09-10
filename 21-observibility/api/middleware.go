package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"

	"sre-labs/21-observibility/internal/obs"
)

// statusRecorder captures the response code so the middleware can measure
// it after the handler returns.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// patternKey carries a *routeHolder through the request context. The
// middleware creates the holder; the innermost withRoute wrapper fills in
// the matched pattern (ServeMux only sets r.Pattern on its own request
// copy, so it never propagates back up on its own).
type patternKey struct{}

type routeHolder struct{ route string }

// routeOfR returns the matched route pattern (e.g. "POST /api/v1/pods")
// recorded by withRoute into the holder stored under ctx; fall back to
// "unmatched" so metrics never get high-cardinality URL labels (404 spam
// with unique paths stays bounded).
//
// It must be called with the enriched ctx (the one handed to the handler
// chain), not the original request — r.WithContext does not flow back up.
func routeOfR(ctx context.Context) string {
	if h, ok := ctx.Value(patternKey{}).(*routeHolder); ok && h.route != "" {
		return h.route
	}
	return "unmatched"
}

// obsMiddleware wires all three pillars into every request:
//  1. traces  — continues an upstream W3C trace (traceparent header) or
//     starts a new root trace, and puts the span into the request context
//  2. metrics — one request counter + latency histogram per method/route
//  3. logs    — a single rich request log line with request_id, trace_id,
//     span_id, method, route, status, duration, client and user agent
func obsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Tracing: join the caller's trace when the W3C traceparent header
		// is present, otherwise start a new root trace.
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		ctx, span, log := obs.StartSpan(ctx, r.Method+" "+r.URL.Path,
			"http.method", r.Method,
			"http.client_ip", clientIP(r),
			"http.user_agent", r.UserAgent(),
		)
		defer obs.EndSpan(span, nil)

		// Request id: generated here (or trust the caller's) and echoed back.
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}
		log = log.With(slog.String("request_id", requestID))
		w.Header().Set("X-Request-ID", requestID)

		// Shared route holder: withRoute (innermost) writes the matched
		// pattern into it; we read it after the handler returns.
		rh := &routeHolder{}
		ctx = context.WithValue(ctx, patternKey{}, rh)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r.WithContext(ctx))

		route := routeOfR(ctx)
		duration := time.Since(start)

		span.SetAttributes(
			attribute.String("http.route", route),
			attribute.Int("http.status_code", rec.status),
			attribute.Float64("http.duration_ms", float64(duration.Microseconds())/1000.0),
		)
		if rec.status >= 500 {
			span.SetStatus(codes.Error, http.StatusText(rec.status))
		}

		// Metrics.
		obs.HTTPRequests.WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
		obs.HTTPDuration.WithLabelValues(r.Method, route).Observe(duration.Seconds())

		// One structured log line per request. trace_id/span_id ride along
		// because `log` came from StartSpan; request_id was added above.
		log.Info("http_request",
			"method", r.Method,
			"route", route,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", float64(duration.Microseconds())/1000.0,
			"remote", clientIP(r),
			"req_bytes", r.ContentLength,
			"ua", r.UserAgent(),
		)
	})
}

// clientIP prefers X-Forwarded-For (behind a proxy) and falls back to the
// peer address.
func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		return xf
	}
	return r.RemoteAddr
}

// newRequestID returns 16 random hex bytes, or "-" if the system RNG fails.
func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "-"
	}
	return hex.EncodeToString(b[:])
}
