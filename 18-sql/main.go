// Learning SQL in Go: create, insert, query, update, transact.
// Run: go run .
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, no server needed
)

func main() {
	db, err := sql.Open("sqlite", "file:sre.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(24 * 60 * time.Minute)

	// 1. Create tables
	createTables(db)

	// 2. Insert rows
	insertServer(db, "web-01", "eu-central", 23.5)
	insertServer(db, "db-01", "us-east", 84.2)
	insertIncident(db, 2, "disk latency high")

	ctx := context.Background()

	// 3. Query one row
	var host string
	row := db.QueryRowContext(ctx, `SELECT hostname FROM servers WHERE id = ?`, 2)
	err = row.Scan(&host)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		fmt.Println("no server with id 2")
	case err != nil:
		log.Fatal(err)
	default:
		fmt.Println("server 2:", host)
	}

	// 4. Query many rows
	rows, err := db.QueryContext(ctx, `SELECT hostname, cpu_pct FROM servers ORDER BY cpu_pct DESC`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var h string
		var cpu float64
		if err := rows.Scan(&h, &cpu); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %.1f%%\n", h, cpu)
	}

	// 5. Update
	res, err := db.Exec(`UPDATE servers SET status = 'degraded' WHERE cpu_pct > 80`)
	if err != nil {
		log.Fatal(err)
	}
	n, _ := res.RowsAffected()
	fmt.Println("degraded servers:", n)

	// 6. Transaction: both statements or none
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE servers SET status = 'incident' WHERE id = 2`); err != nil {
		log.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO incidents (server_id, severity, description) VALUES (2, 'critical', 'cpu on fire')`); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("transaction committed")
}

func createTables(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS servers (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			hostname TEXT NOT NULL UNIQUE,
			region   TEXT NOT NULL,
			status   TEXT NOT NULL DEFAULT 'healthy',
			cpu_pct  REAL NOT NULL DEFAULT 0
		)`)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS incidents (
			id        INTEGER PRIMARY KEY AUTOINCREMENT,
			server_id INTEGER NOT NULL REFERENCES servers(id),
			severity  TEXT NOT NULL,
			description TEXT NOT NULL
		)`)
	if err != nil {
		log.Fatal(err)
	}
}

func insertServer(db *sql.DB, host, region string, cpu float64) {
	_, err := db.Exec(`INSERT INTO servers (hostname, region, cpu_pct) VALUES (?, ?, ?)`, host, region, cpu)
	if err != nil {
		log.Fatal(err)
	}
}

func insertIncident(db *sql.DB, serverID int, description string) {
	_, err := db.Exec(`INSERT INTO incidents (server_id, severity, description) VALUES (?, 'warning', ?)`, serverID, description)
	if err != nil {
		log.Fatal(err)
	}
}
