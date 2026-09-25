# logan — final project: a log ingestion & analysis server

You have built the pieces all course long: HTTP servers, parsers, goroutines,
channels, worker pools, SQLite, config loading, CLI tools. Now you assemble
them into one real service.

**logan** is a server that ingests log lines over HTTP, parses them in
parallel with a worker pool, normalizes error messages into deduplicated
patterns, stores everything in SQLite, evaluates per-service error rates and
opens/resolves alerts — with backpressure when it is overloaded and zero
data loss when you press Ctrl+C.

The skeleton in this directory **compiles, vets and tests green**. Every
place where *you* must write code is marked `TODO(student)` — types,
interfaces, wiring, route tables and contracts are already in place. Your job
is to make it work, not to design it.

```
                    ┌────────────────────────── logan serve ──────────────────────────┐
                    │                                                                 │
 POST /ingest ──────►  echo handler ──► buffered channel ──► worker pool (N workers)   │
   raw log lines     │  (429 when full)      (queue_size)      parse → Signature        │
                    │                                           upsert pattern          │
                    │                                           batch insert            │
                    │                                                │                   │
                    │   alert evaluator (every check_interval) ──►   ▼                   │
                    │   rate > threshold → open alert          SQLite (4 tables)         │
                    │   rate back to normal → resolve alert          │                   │
                    │                                                ▼                   │
 GET /stats /patterns /alerts (/metrics bonus) ◄──────────────── query endpoints        │
                    └──────────────────────────────────────────────────────────────────┘
```

## How the curriculum maps onto this project

Every lab you did exists somewhere in logan. Use this table when you are
stuck: it tells you which lab to reopen.

| Lab | Topic | Where it lives in logan |
|-----|-------|-------------------------|
| 01-server-health-check | HTTP client, maps, status | `logan stats` mindset; testing the API with a client |
| 02-alert-handler | alert shapes, thresholds | `internal/alert` — what an alert is and when it fires |
| 03-zero-value | zero values, initialization | `config.Defaults()`, initializing maps/slices in parser & pool |
| 04-nil-pointer | pointers vs values, nil safety | `store.Alert.ResolvedAt *time.Time` — nil means "still open" |
| 05-log-analyzer | parsing logs, counting levels | `internal/parser` — the core of the project |
| 06-http-server | net/http fundamentals | what echo does under the hood; request/response model |
| 07-struct | structs, tags, composition | every model: `Config`, `Entry`, `Pattern`, `Alert` |
| 08-interface | interfaces, decoupling | `parser.Parser`, `store.Store`, `alert.Evaluator` |
| 09-error-handling | sentinel errors, wrapping, errors.Is | `parser.ErrMalformed`, `pool.ErrQueueFull` |
| 10-json | encoding/json | JSON log lines in, JSON API responses out |
| 11-echo | echo server, routes, handlers | `internal/server` — the whole API |
| 12-time | time.Time, Duration, tickers | log timestamps, alert `Window`/`Cooldown`, batch `FlushInterval` |
| 13-goroutine | go statement, lifetimes | pool workers, evaluator loop |
| 14-channels | channels, select | the ingest queue, worker coordination, ticker select |
| 15-context | context, cancellation | ctx through store calls, graceful shutdown end to end |
| 16-data-racing | races, mutex, atomic | pool counters, shared queue state; `go test -race` must pass |
| 17-worker-pool | the worker pool pattern | `internal/pool` — the heart of the project |
| 18-sql | database/sql, transactions | `internal/store` — schema, upserts, batched inserts |
| 19-k8s-pod-manager | structured code, client apps | (bonus) run logan as a Deployment — see `21-observibility` |
| 20-strings | strings, regexp | `parser.Signature` — normalizing messages into patterns |
| 21-observibility | Prometheus metrics | (bonus) `internal/obs` + `GET /metrics` |
| 22-cli | urfave/cli commands & flags | `cmd/logan` — `serve`, `import`, `stats` |
| 23-config | koanf YAML + env overrides | `internal/config` — `config.yaml` + `LOGAN_*` vars |

## Log formats

Two formats arrive interleaved. A line is **one** of these, or garbage
(garbage is counted, never fatal).

**Plain text**

```
2026-09-10T12:00:00Z ERROR [billing] payment failed for order 12345
└─ RFC3339 timestamp     level  service  free-text message
```

**JSON** (one object per line)

```json
{"ts":"2026-09-10T12:00:00Z","level":"error","service":"billing","msg":"payment failed for order 12345"}
```

Levels normalize to `DEBUG | INFO | WARN | ERROR` regardless of input casing
(`error`, `Error`, `err` are all `ERROR`).

## Pattern signatures

`parser.Signature` collapses volatile parts of a message so millions of
variations collapse into a handful of patterns:

| Input message | Signature |
|---|---|
| `payment failed for order 12345` | `payment failed for order <NUM>` |
| `timeout connecting to 10.0.0.7` | `timeout connecting to <IP>` |
| `session 3f2b8c9e-… expired` | `session <UUID> expired` |
| `config "prod" ignored` | `config <STR> ignored` |

Patterns are deduplicated **per service** (`UNIQUE(service_id, signature)`)
with running `occurrences` and `last_seen` — one upsert.

## Database schema (fixed — do not redesign)

Four SQLite tables, created idempotently (`CREATE TABLE IF NOT EXISTS`) in
`store.Migrate`. The graders check columns, constraints and indexes by name.

1. **services** — `(id INTEGER PK, name TEXT NOT NULL UNIQUE)`
2. **patterns** — `(id INTEGER PK, service_id INTEGER NOT NULL REFERENCES services(id), signature TEXT NOT NULL, occurrences INTEGER NOT NULL DEFAULT 1, first_seen NOT NULL, last_seen NOT NULL, UNIQUE(service_id, signature))`
3. **log_entries** — `(id INTEGER PK, service_id INTEGER NOT NULL REFERENCES services(id), level TEXT NOT NULL, message TEXT NOT NULL, pattern_id INTEGER NULL REFERENCES patterns(id), ts NOT NULL)` — insert-heavy: written in **batched transactions**, indexed on `(service_id, ts)` and `(service_id, level)`
4. **alerts** — `(id INTEGER PK, service_id INTEGER NOT NULL REFERENCES services(id), pattern_id INTEGER NULL REFERENCES patterns(id), rate REAL NOT NULL, threshold REAL NOT NULL, window_seconds INTEGER NOT NULL, created_at NOT NULL, resolved_at NULL)` — `resolved_at IS NULL` means the alert is open

Driver: `modernc.org/sqlite` (pure Go — no cgo, so `-race` is painless).

## HTTP API (fixed contract)

| Method & path | Purpose | Success | Overloaded |
|---|---|---|---|
| `POST /ingest` | feed raw lines (body = newline-delimited log lines, `text/plain`) | `202` `{"status":"queued","lines":N,"rejected":0}` | `429` `{"error":"queue_full","rejected":N}` when **every** line was refused |
| `GET /stats` | all-time per-service totals | `{"services":[{"name":"billing","total":1200,"errors":33,"error_rate":0.0275}]}` | — |
| `GET /patterns?top=10&service=billing` | most frequent patterns (`top` default 10, cap 100; `service` optional) | `[{"signature":"payment failed for order <NUM>","occurrences":57,"last_seen":"..."}]` | — |
| `GET /alerts?open=true` | alerts, newest first (`open=true` → only unresolved) | `[{"id":1,"service_id":2,"pattern_id":null,"rate":0.42,"threshold":0.3,"window_seconds":300,"created_at":"...","resolved_at":null}]` | — |
| `GET /metrics` *(bonus)* | Prometheus format | — | — |

Backpressure is part of the grade: `/ingest` **never blocks**. When the
ingest channel is full, lines are refused with `pool.ErrQueueFull` and
refusals are counted. Partial acceptance is allowed (some queued, some
rejected → still `202` with both counts).

## CLI (implemented for you — it breaks loudly until the internals exist)

```
logan serve              # HTTP API + worker pool + alert evaluator
logan import <path>      # bulk-load a file or a directory of *.log, print summary
logan stats              # print the per-service table from the DB

logan -c other.yaml serve            # custom config file
LOGAN_POOL__WORKERS=8 logan serve    # env override beats the file
```

`import` and `serve` both shut down gracefully — Ctrl+C mid-import must lose
zero lines and still print the summary.

## Configuration

`config.yaml` (defaults also encoded in `config.Defaults()`), overridable by
`LOGAN_` env vars where `__` nests (lab 23):

| Key | Default | Meaning |
|---|---|---|
| `server.addr` | `:8080` | HTTP listen address |
| `server.shutdown_grace` | `10s` | max wait for in-flight HTTP requests on shutdown |
| `database.path` | `logan.db` | SQLite file |
| `pool.workers` | `4` | parser goroutines |
| `pool.queue_size` | `1024` | buffered channel capacity (full → 429) |
| `pool.batch_size` | `100` | entries per insert transaction |
| `pool.flush_interval` | `2s` | max wait before a partial batch flushes |
| `alerts.window` | `5m` | error-rate evaluation window |
| `alerts.threshold` | `0.30` | error rate that opens an alert |
| `alerts.cooldown` | `10m` | min time between alerts for the same service+pattern |
| `alerts.check_interval` | `30s` | evaluator period |
| `log.level` | `info` | app log verbosity |

## Repository layout — what is given, what is yours

```
cmd/logan/main.go          GIVEN: full CLI wiring, incl. shutdown order
config.yaml                GIVEN
testdata/sample.log        GIVEN: mixed formats + a corrupt line
internal/config/config.go  structs GIVEN → implement Load()            (lab 23)
internal/parser/parser.go  types & interface GIVEN → implement parsers,
                           ParseLevel, Signature                       (labs 05, 09, 20)
internal/store/store.go    models & Store interface GIVEN → implement
                           Open/Migrate/SQLiteStore (all 11 methods)   (lab 18)
internal/pool/pool.go      contract GIVEN → implement New/Start/worker/
                           Submit/Shutdown/Stats                       (labs 13–17)
internal/alert/alert.go    interface GIVEN → implement Evaluate
                           (Run is given)                              (labs 02, 12)
internal/server/server.go  routes GIVEN → implement the 4 handlers     (labs 10, 11)
internal/obs/obs.go        (bonus) metric names GIVEN → implement New  (lab 21)
internal/parser/parser_test.go  skeleton GIVEN → unskip, extend
```

Every `TODO(student)` comment states its function's contract. Read the
comment, write the body — no design decisions needed beyond what the comment
and this README pin down. `grep -rn "TODO(student)" internal/` is your
backlog.

## Milestones — commit at each stage

Dependencies are added as you need them; after adding the first import of a
new module run:

```sh
go get github.com/knadh/koanf/v2 github.com/knadh/koanf/providers/file \
      github.com/knadh/koanf/providers/env github.com/knadh/koanf/parsers/yaml
go get modernc.org/sqlite
```

**Stage 1 — sequential spine** *(labs 05, 09, 18, 20, 23)*
`config.Load` · `parser` (both formats, levels, `Signature`) · `store`
(Open/Migrate/SQLiteStore). Prove it: `logan import testdata/sample.log`
then `logan stats` shows per-service counts with the corrupt line counted as
malformed.

**Stage 2 — concurrency** *(labs 13–17, 15, 16)*
`pool.New/Start/worker/Submit/Shutdown`. Prove it: import
`testdata/big.log` (≥10k lines, `make demo-data`) is visibly faster with 4
workers than with 1, `go test -race ./...` is clean, and Ctrl+C mid-import
drains and reports `rejected=0` with no lost lines.

**Stage 3 — alerts** *(labs 02, 12)*
`alert.Evaluate`: windowed error rates, open above threshold (cooldown
respected), resolve back below. Prove it: `import` a file whose checkout
lines are >30% errors, then `logan serve` + `GET /alerts?open=true` shows an
open alert; after the window passes with clean traffic, it resolves.

**Stage 4 — the API** *(labs 10, 11)*
the four echo handlers. Prove it: `POST /ingest` a handful of lines (mixed
formats + one corrupt), then `/stats`, `/patterns?top=3`, `/alerts`; hammer
`/ingest` with `queue_size: 8` and watch 429s appear while the server stays
alive.

**Bonus** *(labs 21, 19)* — `obs.New` + `GET /metrics`, and/or a Deployment
manifest to run logan in Kubernetes.

## Acceptance checklist

- [ ] `go vet ./...` clean, `go build ./...` green, `go test -race ./...` green
- [ ] the four tables exist with the exact columns/constraints/indexes above
- [ ] `logan import` on a ≥10k-line file: summary shows parsed + malformed counts, `rejected=0`
- [ ] Ctrl+C mid-ingest (serve *or* import) loses zero lines and exits cleanly
- [ ] `/ingest` under load returns 429 when the queue is full — never hangs
- [ ] corrupt lines are counted (`malformed`), never crash a worker
- [ ] alerts open above the threshold, respect cooldown, and resolve
- [ ] config: `config.yaml` values and `LOGAN_*` env overrides both work
- [ ] one commit per milestone with a message saying what works
- [ ] a ≥10k-line sample file for the demo (`make demo-data` generates one)

## Rules

- Allowed dependencies: the standard library, `echo/v5`, `urfave/cli/v2`,
  `koanf`, `modernc.org/sqlite`, and (bonus) `prometheus/client_golang`.
  Nothing else without asking.
- Stream: read request bodies and files line by line — never load a whole
  file into memory.
- The DB schema, API paths/verbs and JSON response keys are contracts; do
  not rename them.
- `internal/` is where the learning is. Do not move logic into `cmd/`.

## The demo (10 minutes)

1. `logan serve` — show the startup log with your pool settings.
2. Stream `testdata/big.log` into `POST /ingest` with curl while the server
   runs; hit `/stats` and `/patterns?top=5` live.
3. Flood `/ingest` (tiny `queue_size`) to demonstrate 429 backpressure.
4. Trigger an alert: import a burst of errors for one service, then
   `GET /alerts?open=true`.
5. Press **Ctrl+C mid-ingest**: walk through the shutdown log — http →
   pool drain — and prove zero lines were lost (`submitted == parsed +
   malformed`, `rejected` only from the deliberate flood).
6. Restart and `GET /stats` again — same numbers. Data survived.

## Getting started

```sh
cd final-project
go build ./...              # green today — keep it green
grep -rn "TODO(student)" internal/   # your backlog, in dependency order
make import-sample          # breaks with a TODO panic — that is stage 1
```

Start at `internal/parser/parser.go`. Good luck.
