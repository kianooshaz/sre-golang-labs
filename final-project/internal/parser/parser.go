// Package parser turns raw log lines into Entries and normalizes messages
// into pattern signatures. This is lab 05 (log formats), lab 09 (sentinel
// errors) and lab 20 (string surgery) rolled into the project's core.
package parser

import (
	"errors"
	"time"
)

// Level is the canonical severity of a log line. Stored as-is in the DB.
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// ErrMalformed is the sentinel (lab 09) for a line no parser can read.
// Return it WRAPPED with context ("line 42: ..."), and let callers match
// with errors.Is. Corrupt lines are counted and skipped by the pool — they
// must never crash a worker.
var ErrMalformed = errors.New("malformed log line")

// Entry is one parsed log line.
type Entry struct {
	Service string    // e.g. "billing"
	Level   Level     // canonical, uppercase
	Message string    // free-text part, after the fixed fields
	Ts      time.Time // when the line was emitted
	Raw     string    // the untouched input line, for debugging
}

// Parser reads one raw line. Two implementations exist — the JSON format and
// the plain-text format (see README -> "Log formats"). ParseLine picks one
// per line via Accepts.
type Parser interface {
	// Parse converts raw into an Entry, or returns an error wrapping
	// ErrMalformed.
	Parse(raw string) (Entry, error)
	// Accepts reports whether this parser claims the line. Must be a
	// cheap check (one or two string ops), not a full parse.
	Accepts(raw string) bool
}

// ParseFunc is the pipeline's parse step: raw line in, Entry out, or an
// error wrapping ErrMalformed.
type ParseFunc func(raw string) (Entry, error)

// ForLine returns the parser that claims raw, or nil if none does.
// TODO(student): try the JSON parser first, then the text one.
func ForLine(raw string) Parser {
	panic("TODO(student): pick the JSON or text parser for this line")
}

// ParseLine is the default ParseFunc the pool uses: dispatch through
// ForLine; a line no parser accepts becomes an ErrMalformed error.
// TODO(student): a three-liner on top of ForLine — don't duplicate parsing.
func ParseLine(raw string) (Entry, error) {
	panic("TODO(student): dispatch to the right parser")
}

// ---------------------------------------------------------------- JSON ----

// JSONParser parses one JSON object per line:
//
//	{"ts":"2026-09-10T12:00:00Z","level":"error","service":"billing","msg":"payment failed"}
type JSONParser struct{}

// NewJSONParser returns the JSON-format parser.
func NewJSONParser() *JSONParser { return &JSONParser{} }

// Parse implements Parser.
// TODO(student): json.Unmarshal into a small shadow struct; pass the level
// through ParseLevel; parse ts with time.Parse(time.RFC3339, ...); reject
// lines missing service or msg; wrap every failure with ErrMalformed.
func (p *JSONParser) Parse(raw string) (Entry, error) {
	panic("TODO(student): parse a JSON log line")
}

// Accepts implements Parser.
// TODO(student): does the line start with '{'? (trim spaces first)
func (p *JSONParser) Accepts(raw string) bool {
	panic("TODO(student): one cheap strings check")
}

// ---------------------------------------------------------------- Text ----

// TextParser parses the plain-text format:
//
//	2026-09-10T12:00:00Z ERROR [billing] payment failed for order 12345
//
// RFC3339 timestamp, space, LEVEL, space, [service], space, message.
type TextParser struct{}

// NewTextParser returns the plain-text parser.
func NewTextParser() *TextParser { return &TextParser{} }

// Parse implements Parser.
// TODO(student): split off timestamp, level and [service]; everything after
// the closing ']' is the message; wrap failures with ErrMalformed.
// strings fields + strings.Cut or SplitN are enough — no regexp needed here.
func (p *TextParser) Parse(raw string) (Entry, error) {
	panic("TODO(student): parse a plain-text log line")
}

// Accepts implements Parser.
// TODO(student): cheap heuristic — e.g. the second field is a known level.
func (p *TextParser) Accepts(raw string) bool {
	panic("TODO(student): one cheap check")
}

// ------------------------------------------------------------- Helpers ----

// ParseLevel normalizes any casing or alias ("Error", "err", "warning") to a
// canonical Level.
// TODO(student): switch over strings.ToLower; unknown level -> error.
func ParseLevel(s string) (Level, error) {
	panic("TODO(student): normalize a level word")
}

// Signature collapses the volatile parts of a message so that "payment
// failed for order 12345" and "payment failed for order 89012" share one
// pattern. Rules (also in the README — test each one):
//
//	UUIDs          -> <UUID>
//	IPv4 addresses -> <IP>
//	numbers        -> <NUM>
//	quoted strings -> <STR>
//
// TODO(student): four regexp.ReplaceAllString calls, most specific first
// (UUID and IP before bare numbers, or the numbers eat the UUID first).
// This is lab 20's moment to shine.
func Signature(message string) string {
	panic("TODO(student): normalize ids/numbers/IPs/quotes to placeholders")
}
