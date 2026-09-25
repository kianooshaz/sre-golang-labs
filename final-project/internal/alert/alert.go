// Package alert evaluates per-service error rates and opens/resolves alerts
// with hysteresis — threshold plus cooldown. Labs 02 (alert shapes), 12
// (windows), 15 (context) and 18 (aggregation queries) at work.
package alert

import (
	"context"
	"log"
	"time"

	"logan/internal/store"
)

// Config is the evaluator's tuning; mirrors config.Config's Alerts section.
type Config struct {
	// Window is how far back each evaluation looks.
	Window time.Duration
	// Threshold is the error rate (0..1) that constitutes a breach.
	Threshold float64
	// Cooldown is the minimum time between two alerts for the same
	// service and pattern.
	Cooldown time.Duration
	// CheckInterval is how often Run evaluates.
	CheckInterval time.Duration
}

// Evaluator decides when alerts open and close.
type Evaluator interface {
	// Evaluate runs one pass over all services: query the windowed
	// counts, compute each service's error rate, open alerts for
	// breaches (respecting Cooldown) and resolve open alerts for
	// services whose rate fell back to or below the threshold.
	Evaluate(ctx context.Context, now time.Time) error
}

// evaluator is the Store-backed implementation.
// TODO(student): fields — the store and the config. Nothing else needed.
type evaluator struct{}

// NewEvaluator returns an Evaluator backed by st.
// TODO(student)
func NewEvaluator(st store.Store, cfg Config) Evaluator {
	panic("TODO(student): construct the evaluator")
}

// Evaluate implements Evaluator. One pass:
//
//	since := now.Add(-cfg.Window)
//	counts := st.CountsSince(ctx, since)
//	for each service:
//	    rate := float64(errors) / float64(total)   // define 0/0 = 0
//	    if rate > cfg.Threshold:
//	        patternID := the window's most frequent error pattern, or nil
//	        if !st.RecentAlert(ctx, serviceID, patternID, now.Add(-cfg.Cooldown)):
//	            st.OpenAlert(ctx, store.Alert{Rate: rate, Threshold: ..., ...})
//	    else:
//	        st.ResolveOpenAlerts(ctx, serviceID, now)
//
// Notes: honor ctx cancellation between services (shutdown mid-pass is
// normal); a closed DB or other store error returns the error — Run logs it
// and retries on the next tick; a service with total == 0 has rate 0, so it
// resolves, never alerts.
// TODO(student)
func (e *evaluator) Evaluate(ctx context.Context, now time.Time) error {
	panic("TODO(student): one evaluation pass")
}

// Run drives Evaluate every interval until ctx is done (labs 12 + 15:
// time.NewTicker + ctx.Done). This part is implemented — serve calls it in a
// goroutine. Evaluation errors are logged, never fatal; the next tick
// retries.
func Run(ctx context.Context, ev Evaluator, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			if err := ev.Evaluate(ctx, now); err != nil {
				log.Printf("alert: evaluation failed: %v", err)
			}
		}
	}
}
