# 19 — Kubernetes Pod Manager (HTTP API + Worker Pool)

An SRE lab that combines several previous topics (HTTP server, goroutines, channels,
worker pool, race-safe maps) on top of a real local Kubernetes cluster (minikube).

## What it does

```
            POST /api/v1/pods                 every 2 minutes (ticker)
 client ────────────────────────► ┌─────────┐  ┌──────────────────────────────┐
                                  │  API    │  │ Worker pool (2 goroutines)   │
 GET  /api/v1/pods/{name}/status  │         │  │                              │
 client ◄──────────────────────── │         │  │  queue ──► worker ─┐         │
                                  └────┬────┘  │            worker ─┤─► k8s   │
                                       │       └────────▲───────────┘  API    │
                                       ▼                │                     │
                              ┌───────────────────────────────┐    ┌─────────▼──┐
                              │ in-memory store (mutex-guarded)│    │ minikube   │
                              │  pods:     map[name]Pod       │◄───│ pod status │
                              │  statuses: map[name]Status    │    └────────────┘
                              │  queue:    chan string        │
                              └───────────────────────────────┘
```

1. **create pod** (`POST /api/v1/pods`) validates the name, creates the pod in the
   cluster, adds it to the **in-memory pod list** and pushes the name into the work queue.
2. **worker pool** (2 workers in 2 goroutines) consumes the queue. A ticker fires every
   **2 minutes** and re-queues the whole pod list; each worker then reads the pod status
   from Kubernetes and writes it into the **in-memory status map**.
3. **status pod** (`GET /api/v1/pods/{name}/status`) only reads that in-memory map —
   it never calls Kubernetes directly.

All shared state is guarded by a `sync.RWMutex` (run the app with `-race` to show the
class why: `go run -race .`).

## API

| Method | Path                        | Description                          |
|--------|-----------------------------|--------------------------------------|
| GET    | `/healthz`                  | health probe                         |
| POST   | `/api/v1/pods`              | create a pod (`{"name":"nginx-1","image":"nginx:alpine"}`, image optional, default `nginx:alpine`) |
| GET    | `/api/v1/pods`              | list tracked pods                    |
| GET    | `/api/v1/pods/{name}/status`| latest status from the in-memory map |

## Quickstart

```bash
make k8s-start    # start minikube (profile: sre-lab) — needs Docker Desktop running
make k8s-deploy   # create pod-manager namespace + 2 demo pods (demo-web, demo-cache)
make run          # start the app on :8080

# in another terminal
curl -X POST localhost:8080/api/v1/pods -d '{"name":"nginx-1"}'
sleep 3
curl localhost:8080/api/v1/pods/nginx-1/status
```

Other targets: `make build`, `make vet`, `make test`, `make k8s-status`, `make k8s-pods`,
`make k8s-logs`, `make clean`, `make clean-all` (deletes namespace + cluster).

## Layout

```
main.go              entry point
api/                 HTTP server: routing, handlers, logging middleware
internal/store/      in-memory pod list + status map + work queue (mutex-guarded)
internal/kube/       thin client-go wrapper (create pod, get pod, ensure namespace)
internal/worker/     worker pool: 2 goroutines, 2-minute ticker
k8s/                 namespace + demo pod manifests
Makefile             cluster + app lifecycle
```

## Class discussion points

- Why does the status API read from memory instead of the cluster? (decoupling, latency,
  the worker is the only writer — a single-writer pattern)
- Why `sync.RWMutex` instead of `sync.Mutex`? (many concurrent status reads)
- Why is `queue` a channel while the maps need a mutex? (channels for hand-off, mutex
  for shared state)
- What happens if a worker is busy for more than one tick? (queue buffered 128, names
  are dropped if full — a good place to discuss back-pressure)
- What happens if the app restarts? (in-memory state is lost; the demo pods still run —
  a good segue into reconciliation/StatefulSets)
