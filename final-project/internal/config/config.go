// Package config holds logan's typed configuration and the loader that turns
// config.yaml + LOGAN_* environment variables into it.
//
// The structs below are the contract — do not change field names or koanf
// tags: config.yaml, the README and the graders depend on them. Your job in
// this package is Load (and only Load).
package config

import "time"

// Config is logan's full configuration tree. The koanf tags define the YAML
// keys; every key is overridable through an environment variable using `__`
// as the nesting separator, e.g. LOGAN_POOL__WORKERS overrides pool.workers.
type Config struct {
	Server struct {
		// Addr is the HTTP listen address, e.g. ":8080".
		Addr string `koanf:"addr"`
		// ShutdownGrace is how long graceful shutdown may take before
		// in-flight requests are cut off.
		ShutdownGrace time.Duration `koanf:"shutdown_grace"`
	} `koanf:"server"`

	Database struct {
		// Path is the SQLite file path.
		Path string `koanf:"path"`
	} `koanf:"database"`

	Pool struct {
		// Workers is the number of parser goroutines.
		Workers int `koanf:"workers"`
		// QueueSize is the buffered-channel capacity between the HTTP
		// handler and the workers. When it is full, /ingest answers 429.
		QueueSize int `koanf:"queue_size"`
		// BatchSize is how many entries a worker accumulates before
		// flushing them to the store in one transaction.
		BatchSize int `koanf:"batch_size"`
		// FlushInterval is how long a worker waits at most before
		// flushing a partially-filled batch.
		FlushInterval time.Duration `koanf:"flush_interval"`
	} `koanf:"pool"`

	Alerts struct {
		// Window is the sliding window over which the per-service error
		// rate is computed.
		Window time.Duration `koanf:"window"`
		// Threshold is the error rate (0..1) at which an alert opens.
		Threshold float64 `koanf:"threshold"`
		// Cooldown is the minimum time between two alerts for the same
		// service and pattern.
		Cooldown time.Duration `koanf:"cooldown"`
		// CheckInterval is how often the evaluator runs.
		CheckInterval time.Duration `koanf:"check_interval"`
	} `koanf:"alerts"`

	Log struct {
		// Level is the app's own log verbosity ("debug"|"info"|"warn"|"error").
		Level string `koanf:"level"`
	} `koanf:"log"`
}

// Defaults returns the values config.yaml ships with. Load must start from
// here so a partial config file still yields a complete Config.
func Defaults() Config {
	var c Config
	c.Server.Addr = ":8080"
	c.Server.ShutdownGrace = 10 * time.Second
	c.Database.Path = "logan.db"
	c.Pool.Workers = 4
	c.Pool.QueueSize = 1024
	c.Pool.BatchSize = 100
	c.Pool.FlushInterval = 2 * time.Second
	c.Alerts.Window = 5 * time.Minute
	c.Alerts.Threshold = 0.30
	c.Alerts.Cooldown = 10 * time.Minute
	c.Alerts.CheckInterval = 30 * time.Second
	c.Log.Level = "info"
	return c
}

// Load builds the final Config by layering, in order of increasing
// precedence:
//
//  1. Defaults()
//  2. the YAML file at path (a missing file is fine when env vars provide
//     the values; a malformed file is an error)
//  3. environment variables prefixed with LOGAN_ (LOGAN_POOL__WORKERS ->
//     pool.workers; `__` in the env name separates nesting levels)
//
// Validate the result lightly: workers >= 1, queue_size >= 1, batch_size >= 1
// and threshold in (0,1]. Everything else is the caller's problem.
//
// TODO(student): implement with koanf, exactly like lab 23:
//   - koanf.New(".")
//   - start from Defaults() so unmarshal merges over the defaults
//   - file.Provider(path) + yaml.Parser()
//   - env.Provider("LOGAN_", ".", callback) mapping `__` -> `.` and
//     lower-casing the rest
//   - k.Unmarshal("", &cfg)
func Load(path string) (Config, error) {
	panic("TODO(student): layer defaults + YAML + LOGAN_* env into a Config")
}
