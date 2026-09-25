// logan — log ingestion and analysis server. Final-project entry point.
//
// This wiring is complete on purpose: read it top to bottom to see how the
// pieces fit, then follow the TODO(student) markers into internal/. Nothing
// runs until you implement those — every entry point panics with a TODO.
//
//	logan serve            HTTP API + worker pool + alert evaluator
//	logan import <path>    bulk-load a log file or a directory of *.log
//	logan stats            print per-service totals from the database
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"

	"logan/internal/alert"
	"logan/internal/config"
	"logan/internal/obs"
	"logan/internal/parser"
	"logan/internal/pool"
	"logan/internal/server"
	"logan/internal/store"
)

func main() {
	app := &cli.App{
		Name:  "logan",
		Usage: "ingest, analyze and alert on service logs",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.yaml",
				Usage:   "path to the YAML config file",
			},
		},
		Commands: []*cli.Command{
			serveCmd(),
			importCmd(),
			statsCmd(),
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// serveCmd is the full pipeline. The shutdown order matters and is graded:
// HTTP first (stop accepting), then the pool drain (zero lines lost), then
// the evaluator, then the database.
func serveCmd() *cli.Command {
	return &cli.Command{
		Name:  "serve",
		Usage: "start the HTTP API, the worker pool and the alert evaluator",
		Action: func(c *cli.Context) error {
			cfg, err := config.Load(c.String("config"))
			if err != nil {
				return err
			}

			// Ctrl+C / SIGTERM anywhere in the process cancels ctx.
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			db, err := store.Open(cfg.Database.Path)
			if err != nil {
				return err
			}
			defer db.Close()

			if err := store.Migrate(ctx, db); err != nil {
				return err
			}
			st, err := store.NewSQLite(db)
			if err != nil {
				return err
			}

			pl := pool.New(cfg.Pool.Workers, cfg.Pool.QueueSize, cfg.Pool.BatchSize,
				cfg.Pool.FlushInterval, parser.ParseLine, st)
			ev := alert.NewEvaluator(st, alert.Config{
				Window:        cfg.Alerts.Window,
				Threshold:     cfg.Alerts.Threshold,
				Cooldown:      cfg.Alerts.Cooldown,
				CheckInterval: cfg.Alerts.CheckInterval,
			})

			var m *obs.Metrics // TODO(student bonus): m = obs.New()
			srv := server.New(cfg, pl, st, ev, m)

			pl.Start(ctx)
			go alert.Run(ctx, ev, cfg.Alerts.CheckInterval)

			log.Printf("logan listening on %s (workers=%d queue=%d)",
				cfg.Server.Addr, cfg.Pool.Workers, cfg.Pool.QueueSize)

			// Blocks until Ctrl+C/SIGTERM; echo drains in-flight requests
			// for up to Server.ShutdownGrace before returning.
			if err := srv.Start(ctx); err != nil {
				log.Printf("http server: %v", err)
			}

			log.Println("http stopped; draining worker pool")
			if err := pl.Shutdown(context.Background()); err != nil {
				log.Printf("pool shutdown: %v", err)
			}
			if s := pl.Stats(); s.Malformed > 0 || s.Rejected > 0 {
				log.Printf("pool totals: submitted=%d parsed=%d malformed=%d rejected=%d",
					s.Submitted, s.Parsed, s.Malformed, s.Rejected)
			}
			return nil
		},
	}
}

// importCmd bulk-feeds files through the same pool + store, without HTTP.
// It never drops lines: when the queue is full it waits, and Ctrl+C stops
// reading, drains, and still prints the summary.
func importCmd() *cli.Command {
	return &cli.Command{
		Name:      "import",
		Usage:     "bulk-import a log file or a directory of *.log files",
		ArgsUsage: "<file-or-dir>",
		Action: func(c *cli.Context) error {
			path := c.Args().First()
			if path == "" {
				return cli.Exit("usage: logan import <file-or-dir>", 2)
			}
			cfg, err := config.Load(c.String("config"))
			if err != nil {
				return err
			}
			files, err := walkLogs(path)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				return fmt.Errorf("no .log files under %s", path)
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			db, err := store.Open(cfg.Database.Path)
			if err != nil {
				return err
			}
			defer db.Close()
			if err := store.Migrate(ctx, db); err != nil {
				return err
			}
			st, err := store.NewSQLite(db)
			if err != nil {
				return err
			}

			pl := pool.New(cfg.Pool.Workers, cfg.Pool.QueueSize, cfg.Pool.BatchSize,
				cfg.Pool.FlushInterval, parser.ParseLine, st)
			pl.Start(ctx)

			started := time.Now()
			for _, f := range files {
				if err := feed(ctx, pl, f); err != nil {
					return err
				}
			}
			if err := pl.Shutdown(context.Background()); err != nil {
				return err
			}

			s := pl.Stats()
			log.Printf("imported %d file(s) in %s: submitted=%d parsed=%d malformed=%d rejected=%d",
				len(files), time.Since(started).Round(time.Millisecond),
				s.Submitted, s.Parsed, s.Malformed, s.Rejected)
			return nil
		},
	}
}

// statsCmd prints the per-service table straight from the database.
func statsCmd() *cli.Command {
	return &cli.Command{
		Name:  "stats",
		Usage: "print per-service totals and error rates from the database",
		Action: func(c *cli.Context) error {
			cfg, err := config.Load(c.String("config"))
			if err != nil {
				return err
			}
			ctx := context.Background()

			db, err := store.Open(cfg.Database.Path)
			if err != nil {
				return err
			}
			defer db.Close()

			st, err := store.NewSQLite(db)
			if err != nil {
				return err
			}
			rows, err := st.Stats(ctx)
			if err != nil {
				return err
			}

			fmt.Printf("%-20s %12s %12s %10s\n", "SERVICE", "TOTAL", "ERRORS", "RATE")
			for _, r := range rows {
				fmt.Printf("%-20s %12d %12d %9.2f%%\n", r.Name, r.Total, r.Errors, r.ErrorRate*100)
			}
			return nil
		},
	}
}

// feed streams one file into the pool line by line. On a full queue it waits
// and retries the same line — a bulk import may slow down, never lose data.
func feed(ctx context.Context, pl *pool.Pool, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		for {
			err := pl.Submit(line)
			if err == nil {
				break
			}
			if !errors.Is(err, pool.ErrQueueFull) {
				return err
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Millisecond):
			}
		}
	}
	return sc.Err()
}

// walkLogs turns the argument into a list of .log files: a directory is
// globbed one level deep, a file is taken as-is.
func walkLogs(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	return filepath.Glob(filepath.Join(path, "*.log"))
}
