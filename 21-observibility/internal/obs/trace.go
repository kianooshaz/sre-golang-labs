package obs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is the tracer the app's packages use to start spans.
var Tracer trace.Tracer = otel.Tracer("sre-labs/pod-manager")

// initTracing configures the global tracer provider. When OTLPEndpoint is
// empty, spans are still created (so trace_id lands in the logs) but are
// dropped instead of exported — useful for local development.
func initTracing(ctx context.Context, cfg Config) error {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVersion),
			attribute.String("deployment.environment.name", cfg.Env),
		),
	)
	if err != nil {
		return fmt.Errorf("build trace resource: %w", err)
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		// Sample everything; flip to ParentBased(TraceIDRatioBased(0.1)) in prod.
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	}

	if cfg.OTLPEndpoint != "" {
		exp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(), // local collector; use TLS in prod
		)
		if err != nil {
			return fmt.Errorf("create otlp exporter: %w", err)
		}
		bsp := sdktrace.NewBatchSpanProcessor(exp,
			sdktrace.WithBatchTimeout(2*time.Second),
			sdktrace.WithMaxQueueSize(2048),
		)
		opts = append(opts, sdktrace.WithSpanProcessor(bsp))
		state.shutdownFns = append(state.shutdownFns, func(shutdownCtx context.Context) error {
			return bsp.Shutdown(shutdownCtx)
		})
	}
	// Without an exporter, the provider still creates valid span contexts
	// (trace_id in every log line); ended spans are simply not exported.

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	Tracer = tp.Tracer("sre-labs/pod-manager")

	// W3C traceparent/tracestate: lets incoming HTTP headers continue a trace
	// started upstream, and lets us propagate into outbound calls.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// StartSpan starts a named span and returns the updated context, the span
// and a ready-to-use logger that already carries trace_id and span_id.
// It is the single helper every package uses so traces and logs always agree.
func StartSpan(ctx context.Context, name string, attrs ...any) (context.Context, trace.Span, *slog.Logger) {
	ctx, span := Tracer.Start(ctx, name)
	if len(attrs) > 0 {
		span.SetAttributes(Attrs(attrs)...)
	}
	return ctx, span, SpanLogger(ctx)
}

// SpanLogger returns the default logger enriched with the trace context of
// the current span, so every log line can be joined to its trace.
func SpanLogger(ctx context.Context) *slog.Logger {
	log := Logger()
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return log
	}
	return log.With(
		slog.String("trace_id", sc.TraceID().String()),
		slog.String("span_id", sc.SpanID().String()),
	)
}

// DetachSpanContext returns a snapshot of the current span context, safe to
// store and reuse later from a different goroutine — e.g. the worker pool
// links its status-check spans back to the trace that created the pod.
func DetachSpanContext(ctx context.Context) (trace.SpanContext, bool) {
	sc := trace.SpanContextFromContext(ctx)
	return sc, sc.IsValid()
}

// StartLinkedSpan starts a new (detached) span in the current goroutine and
// links it to a previously captured span context. The link shows up in the
// trace UI as a cross-reference, connecting async work to its origin.
func StartLinkedSpan(ctx context.Context, linked trace.SpanContext, name string, attrs ...any) (context.Context, trace.Span, *slog.Logger) {
	var links []trace.Link
	if linked.IsValid() {
		links = []trace.Link{{SpanContext: linked}}
	}
	ctx, span := Tracer.Start(ctx, name, trace.WithLinks(links...))
	if len(attrs) > 0 {
		span.SetAttributes(AttrsOf(attrs...)...)
	}
	return ctx, span, SpanLogger(ctx)
}

// EndSpan finishes a span and marks it failed when err is non-nil,
// so backend UIs (Tempo/Jaeger/Zipkin) show red spans for failed work.
func EndSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(1, err.Error()) // codes.Error
	}
	span.End()
}

// Attrs converts slog-style key/value pairs into OTel span attributes.
// Values that are not one of the basic kinds become their fmt %v rendering.
func Attrs(kv []any) []attribute.KeyValue {
	return AttrsOf(kv...)
}

// AttrsOf is the variadic form of Attrs: AttrsOf("pod", name, "ok", true).
func AttrsOf(kv ...any) []attribute.KeyValue {
	if len(kv)%2 != 0 {
		kv = append(kv, nil)
	}
	out := make([]attribute.KeyValue, 0, len(kv)/2)
	for i := 0; i < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", kv[i])
		}
		out = append(out, anyToAttr(key, kv[i+1]))
	}
	return out
}

func anyToAttr(key string, v any) attribute.KeyValue {
	switch val := v.(type) {
	case string:
		return attribute.String(key, val)
	case int:
		return attribute.Int(key, val)
	case int64:
		return attribute.Int64(key, val)
	case float64:
		return attribute.Float64(key, val)
	case bool:
		return attribute.Bool(key, val)
	case error:
		return attribute.String(key, val.Error())
	case time.Duration:
		return attribute.String(key, val.String())
	default:
		return attribute.String(key, fmt.Sprintf("%v", val))
	}
}
