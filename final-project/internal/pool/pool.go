// Package pool is logan's heart: a fixed set of worker goroutines (labs 13,
// 14, 17) that pull raw lines off one buffered channel, parse them, link them
// to patterns and store them in batches — race-free (lab 16), with
// backpressure when the queue fills up, and a zero-line-loss graceful drain
// on shutdown (lab 15).
package pool

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"logan/internal/parser"
	"logan/internal/store"
)

// ErrQueueFull is returned by Submit when the ingest channel is at capacity.
// The HTTP handler maps it to 429 Too Many Requests — that IS the
// backpressure mechanism working, not a bug.
var ErrQueueFull = errors.New("ingest queue is full")

// Stats is a point-in-time snapshot of the pool's counters. All of them must
// be safe to read while workers run: atomics or a mutex, never a bare int
// written by many goroutines (lab 16).
type Stats struct {
	Submitted uint64 // accepted by Submit
	Parsed    uint64 // turned into entries and stored
	Malformed uint64 // corrupt lines: counted, skipped, never fatal
	Rejected  uint64 // refused by Submit with ErrQueueFull
}

// Pool is the worker pool. The fields are deliberately yours to design; you
// need at least the jobs channel, the worker count, the batch tuning, the
// parser, the store, the counters and something that makes Submit after
// Shutdown safe.
type Pool struct {
	// TODO(student): unexported fields — channel, config, parser, store,
	// counters, WaitGroup, closed flag / sync.Once.
}

// New builds a Pool: workers goroutines will pull from a QueueSize buffered
// channel; each worker flushes to the store every BatchSize entries or every
// FlushInterval, whichever comes first. New only wires state — it launches
// nothing (that is Start's job).
// TODO(student)
func New(workers, queueSize, batchSize int, flushInterval time.Duration, parse parser.ParseFunc, st store.Store) *Pool {
	panic("TODO(student): construct the pool")
}

// Start launches the workers. ctx is the pool's lifetime: canceling it must
// not drop work already accepted — a worker canceled mid-line still finishes
// that line and its batch before exiting.
// TODO(student): one `go p.worker(ctx)` per worker.
func (p *Pool) Start(ctx context.Context) {
	panic("TODO(student): launch the workers")
}

// worker is the body of one goroutine. Sketch of the loop you are writing:
//
//	for raw := range jobs {
//	    entry, err := parse(raw)
//	    if errors.Is(err, parser.ErrMalformed) { Malformed++; continue }
//	    sig := parser.Signature(entry.Message)
//	    serviceID, _ := st.UpsertService(ctx, entry.Service)
//	    patternID, _ := st.UpsertPattern(ctx, serviceID, sig, entry.Ts)
//	    batch = append(batch, store.Entry{...})
//	    if len(batch) >= batchSize { flush(ctx, batch); batch = batch[:0] }
//	}
//	flush(ctx, batch) // drain the tail
//
// TODO(student): plus the FlushInterval timer (lab 12) so a quiet worker
// still flushes its partial batch, and error handling that counts instead
// of crashing the worker.
func (p *Pool) worker(ctx context.Context) {
	panic("TODO(student): parse -> pattern -> batch -> flush loop")
}

// Submit hands one raw line to the pool WITHOUT blocking. When the queue is
// full it returns ErrQueueFull immediately — the HTTP handler turns that
// into 429. After Shutdown it must also return (ErrQueueFull is fine); a
// Submit that panics or hangs here is a grading bug.
// TODO(student): select with a default case; guard with your closed flag.
func (p *Pool) Submit(raw string) error {
	panic("TODO(student): enqueue or reject — never block")
}

// Shutdown performs the graceful drain. Acceptance-critical — "Ctrl+C
// mid-ingest loses zero lines" (README):
//
//  1. refuse new lines from now on (Submit returns ErrQueueFull)
//  2. close the jobs channel and let workers drain everything already queued
//  3. make sure every partial batch is flushed
//  4. wait for all workers to exit (WaitGroup)
//
// When Shutdown returns, every submitted line is either stored or counted in
// Stats as malformed. It must be safe to call once, from the shutdown path.
// TODO(student)
func (p *Pool) Shutdown(ctx context.Context) error {
	panic("TODO(student): refuse -> close(jobs) -> drain -> flush -> wait")
}

// Stats returns a snapshot of the counters.
// TODO(student): atomic.LoadUint64 each counter.
func (p *Pool) Stats() Stats {
	panic("TODO(student): read the counters")
}

// ensure the atomic package stays imported for your counters.
var _ = atomic.LoadUint64
