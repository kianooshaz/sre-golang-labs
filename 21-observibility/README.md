# 21-observibility

Lab 21: the pod-manager from [19-k8s-pod-manager](../19-k8s-pod-manager) rebuilt
with the three pillars of observability — **metrics**, **span tracing** and
**rich structured logs** — wired through one shared `internal/obs` package.

The app itself is unchanged: an HTTP API that creates pods in a local
minikube cluster via client-go, backed by an in-memory store and a worker
pool that polls pod status.

## What was added

| Pillar | Where | What you get |
|---|---|---|
| Metrics | `internal/obs/metrics.go`, exposed at `GET /metrics` | Prometheus counters/histograms/gauges: `podmanager_http_requests_total`, `podmanager_http_request_duration_seconds`, `podmanager_pod_creations_total{outcome}`, `podmanager_pod_create_duration_seconds`, `podmanager_pod_status_checks_total{worker,outcome}`, `podmanager_workqueue_length`, `podmanager_tracked_pods`, plus Go runtime collectors |
| Traces | `internal/obs/trace.go` | OpenTelemetry spans for every request and every Kubernetes call (`handlers.createPod` → `k8s.create_pod`), W3C `traceparent` extraction/injection, error spans marked red. Export via OTLP gRPC |
| Logs | `internal/obs/obs.go` | slog JSON with base attrs (`service`, `env`, `version`, `pid`) on every line, durations as human strings, and `trace_id`/`span_id`/`request_id` injected so any log line joins to its trace |

Route metrics are labeled with the low-cardinality route pattern
(e.g. `POST /api/v1/pods`), not raw paths, so unique-URL 404 spam cannot
explode label cardinality.

## Endpoints

| Path | Purpose |
|---|---|
| `GET /healthz` | liveness |
| `POST /api/v1/pods` | create a pod (`{"name":"nginx-1","image":"nginx:alpine"}`) |
| `GET /api/v1/pods` | list tracked pods |
| `GET /api/v1/pods/{name}/status` | latest worker-fetched status |
| `GET /metrics` | Prometheus scrape |
| `GET /debug/vars` | expvar runtime gauges (goroutines, heap) |

## Configuration

| Env var | Default | Meaning |
|---|---|---|
| `OTLP_ENDPOINT` | *(empty)* | `host:port` of an OTLP gRPC collector; empty = spans created but not exported (trace_id still in logs) |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `APP_ENV` | `dev` | free-form environment tag on every log line and span |

## Run it

```sh
# 1. Start a local cluster and demo pods (same as lab 19)
make k8s-start k8s-deploy

# 2. Run the app
make run
# or with tracing exported to a local collector (Jaeger all-in-one):
OTLP_ENDPOINT=localhost:4317 make run

# 3. Create a pod and watch the signals
curl -s localhost:8080/api/v1/pods -d '{"name":"nginx-1"}'
curl -s localhost:8080/metrics | grep podmanager
```

Graceful shutdown: SIGINT/SIGTERM drains the HTTP server (10s), stops the
worker pool, then flushes the trace exporter.

## Layout

```
api/            HTTP handlers + observability middleware
internal/obs/   logging, metrics, tracing — the one wiring point
internal/kube/  Kubernetes client, each call wrapped in a span
internal/store/ in-memory state + workqueue (queue gauges/counters)
internal/worker/ worker pool (per-check spans and outcome metrics)
internal/app/   boot order: obs → store → kube → workers → HTTP
k8s/            namespace and demo pod manifests
```
