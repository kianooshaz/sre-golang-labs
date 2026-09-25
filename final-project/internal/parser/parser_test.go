package parser

import (
	"errors"
	"testing"
	"time"
)

// This file is the minimum test bar (README -> Acceptance: `go test -race
// ./...` green). Unskip each test as you implement, then extend the tables —
// table-driven, like the labs. You are expected to add more tests of your
// own (pool drain, store upserts, evaluator cooldown, ...).

func TestParseLevel(t *testing.T) {
	t.Skip("TODO(student): unskip once ParseLevel is implemented")

	cases := []struct {
		in      string
		want    Level
		wantErr bool
	}{
		{"error", LevelError, false},
		// TODO(student): INFO / Warn / debug / err aliases + one invalid.
	}
	for _, tc := range cases {
		got, err := ParseLevel(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseLevel(%q): want error, got %q", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseLevel(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseLevel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSignature(t *testing.T) {
	t.Skip("TODO(student): unskip once Signature is implemented")

	cases := []struct{ msg, want string }{
		{"payment failed for order 12345", "payment failed for order <NUM>"},
		{"timeout connecting to 10.0.0.7", "timeout connecting to <IP>"},
		{"session 3f2b8c9e-1d4a-4b6f-9a2c-5e7d8f1a2b3c expired", "session <UUID> expired"},
		{`config "prod" ignored`, `config <STR> ignored`},
		// TODO(student): a UUID full of digits (replacement order matters!),
		// two numbers in one message, an empty message.
	}
	for _, tc := range cases {
		if got := Signature(tc.msg); got != tc.want {
			t.Errorf("Signature(%q) = %q, want %q", tc.msg, got, tc.want)
		}
	}
}

func TestParseLine(t *testing.T) {
	t.Skip("TODO(student): unskip once ParseLine and the parsers are implemented")

	// JSON line.
	e, err := ParseLine(`{"ts":"2026-09-10T12:00:00Z","level":"error","service":"billing","msg":"boom 42"}`)
	if err != nil {
		t.Fatalf("json line: %v", err)
	}
	if e.Service != "billing" || e.Level != LevelError || e.Message != "boom 42" {
		t.Errorf("json line parsed wrong: %+v", e)
	}
	if e.Ts.After(time.Now()) {
		t.Errorf("json line: timestamp in the future: %v", e.Ts)
	}

	// Text line.
	e, err = ParseLine("2026-09-10T12:00:00Z ERROR [checkout] payment declined for card 7")
	if err != nil {
		t.Fatalf("text line: %v", err)
	}
	if e.Service != "checkout" || e.Level != LevelError {
		t.Errorf("text line parsed wrong: %+v", e)
	}

	// Corrupt line -> ErrMalformed (wrapped), never a panic.
	if _, err := ParseLine("this is not a log line"); !errors.Is(err, ErrMalformed) {
		t.Errorf("corrupt line: want ErrMalformed, got %v", err)
	}
}
