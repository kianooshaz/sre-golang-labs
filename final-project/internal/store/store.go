// Package store persists entries, patterns and alerts in SQLite behind an
// interface (lab 08) so tests can swap it out (lab 18 for the SQL itself).
//
// The column layout is fixed — see README -> "Database schema". Implement
// SQLiteStore with database/sql + the pure-Go driver modernc.org/sqlite.
package store

import (
	"context"
	"database/sql"
	"time"
)

// ------------------------------------------------------------ Row types ----

// Service is one emitting service. Name is UNIQUE.
type Service struct {
	ID   int64
	Name string
}

// Entry is one stored log line, already linked to its service and pattern.
type Entry struct {
	ServiceID int64
	PatternID int64 // 0 means "no pattern" — store as NULL, not 0
	Level     string
	Message   string
	Ts        time.Time
}

// Pattern is a normalized message signature, deduplicated per service.
type Pattern struct {
	ID          int64
	ServiceID   int64
	Signature   string
	Occurrences int64
	FirstSeen   time.Time
	LastSeen    time.Time
}

// PatternStat is one row of the /patterns?top=N report.
type PatternStat struct {
	Signature   string
	Occurrences int64
	LastSeen    time.Time
}

// Alert is an opened — and maybe resolved — error-rate alert.
// ResolvedAt == nil means the alert is still open: a *time.Time, lab 04.
type Alert struct {
	ID         int64
	ServiceID  int64
	PatternID  *int64 // nullable FK
	Rate       float64
	Threshold  float64
	WindowSecs int64
	CreatedAt  time.Time
	ResolvedAt *time.Time
}

// ServiceStats is one row of the /stats report (all-time numbers).
type ServiceStats struct {
	Name      string
	Total     int64
	Errors    int64
	ErrorRate float64 // Errors / Total, 0 when Total is 0
}

// ServiceCounts is one service's windowed counts — the alert evaluator
// works on these.
type ServiceCounts struct {
	ServiceID int64
	Name      string
	Total     int64
	Errors    int64
}

// ------------------------------------------------------------- Contract ----

// Store is everything the rest of logan needs from the database.
type Store interface {
	// UpsertService returns the id of the named service, inserting the
	// row on first sight.
	UpsertService(ctx context.Context, name string) (int64, error)

	// UpsertPattern inserts the (serviceID, signature) pattern or bumps
	// occurrences and last_seen, and returns its id. The UNIQUE(service_id,
	// signature) constraint makes this one INSERT ... ON CONFLICT DO UPDATE.
	UpsertPattern(ctx context.Context, serviceID int64, signature string, seenAt time.Time) (int64, error)

	// InsertEntries stores a batch in ONE transaction. Empty slice = no-op.
	// A failed batch must insert nothing — that is the point of batching.
	InsertEntries(ctx context.Context, entries []Entry) error

	// Stats aggregates all-time totals and error rates per service for
	// GET /stats. Services with zero entries may appear with Total 0.
	Stats(ctx context.Context) ([]ServiceStats, error)

	// CountsSince returns per-service total/error counts for entries with
	// ts >= since — the alert evaluator's window query.
	CountsSince(ctx context.Context, since time.Time) ([]ServiceCounts, error)

	// TopPatterns returns the limit most-frequent patterns, optionally
	// scoped to one service name ("" = all services), most frequent first.
	TopPatterns(ctx context.Context, service string, limit int) ([]PatternStat, error)

	// RecentAlert reports whether service+pattern already has an alert
	// created after since — the evaluator's cooldown check.
	RecentAlert(ctx context.Context, serviceID, patternID int64, since time.Time) (bool, error)

	// OpenAlert records a new alert and returns its id.
	OpenAlert(ctx context.Context, a Alert) (int64, error)

	// Alerts lists alerts, newest first; onlyOpen filters to
	// resolved_at IS NULL.
	Alerts(ctx context.Context, onlyOpen bool) ([]Alert, error)

	// ResolveOpenAlerts marks every open alert of the service as resolved
	// at now and returns how many rows changed.
	ResolveOpenAlerts(ctx context.Context, serviceID int64, now time.Time) (int64, error)

	// Close releases the database handle.
	Close() error
}

// ------------------------------------------------------------ SQLite impl ----

// Open connects to the SQLite database at path. Use sql.Open("sqlite", ...)
// with the pure-Go driver modernc.org/sqlite (no cgo — `go test -race` stays
// happy) and set sane pool limits (see lab 18). It must not touch the
// schema — call Migrate after Open.
// TODO(student)
func Open(path string) (*sql.DB, error) {
	panic("TODO(student): sql.Open + ping + pool limits")
}

// Migrate creates the four tables and their indexes if they do not exist.
// The exact DDL is specified in README -> "Database schema"; required, in
// dependency order:
//
//	services     (id PK, name TEXT NOT NULL UNIQUE)
//	patterns     (id PK, service_id NOT NULL -> services.id,
//	              signature TEXT NOT NULL, occurrences NOT NULL DEFAULT 1,
//	              first_seen NOT NULL, last_seen NOT NULL,
//	              UNIQUE(service_id, signature))
//	log_entries  (id PK, service_id NOT NULL -> services.id,
//	              level TEXT NOT NULL, message TEXT NOT NULL,
//	              pattern_id NULL -> patterns.id, ts NOT NULL)
//	              INDEX (service_id, ts); INDEX (service_id, level)
//	alerts       (id PK, service_id NOT NULL -> services.id,
//	              pattern_id NULL -> patterns.id, rate REAL NOT NULL,
//	              threshold REAL NOT NULL, window_seconds INTEGER NOT NULL,
//	              created_at NOT NULL, resolved_at NULL)
//
// Use CREATE TABLE IF NOT EXISTS so restarts are idempotent.
// TODO(student): write the DDL and Exec it.
func Migrate(ctx context.Context, db *sql.DB) error {
	panic("TODO(student): create services, patterns, log_entries, alerts")
}

// SQLiteStore is the Store implementation over an open *sql.DB.
// TODO(student): give it a db field, then implement every Store method on
// *SQLiteStore. Put each method's SQL right where it belongs — no helpers
// layer, this package is the data-access layer.
type SQLiteStore struct{}

// NewSQLite wraps db in a Store.
// TODO(student)
func NewSQLite(db *sql.DB) (Store, error) {
	panic("TODO(student): construct *SQLiteStore")
}
