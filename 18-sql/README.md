# 18 — SQL in Go

One file, six basics of `database/sql`, using SQLite (no server, pure Go).

```bash
go mod tidy
go run .
```

| Step | What it shows                                   |
| ---- | ----------------------------------------------- |
| 1    | `CREATE TABLE` with `Exec`                      |
| 2    | `INSERT` with `?` placeholders                  |
| 3    | One row: `QueryRow` + `sql.ErrNoRows`           |
| 4    | Many rows: `Query` + `rows.Next`/`Scan`/`Close` |
| 5    | `UPDATE` + `RowsAffected`                       |
| 6    | Transaction: `Begin`/`Commit`/`Rollback`        |

Delete `sre.db` to start fresh.
