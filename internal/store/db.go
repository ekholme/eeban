// Package store is the SQLite persistence layer. Queries are hand-written
// against database/sql; the TUI never imports this package directly.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// DB wraps a *sql.DB connection to the eeban database.
type DB struct {
	*sql.DB
}

// Open opens (creating if needed) the SQLite database at path, applies
// connection pragmas, and runs any pending migrations. path may be a plain
// filesystem path or a full "file:" DSN (used by tests).
func Open(ctx context.Context, path string) (*DB, error) {
	sqldb, err := sql.Open("sqlite", buildDSN(path))
	if err != nil {
		return nil, err
	}

	// Single-user TUI: serialize every access through one connection so writes
	// never hit SQLITE_BUSY. WAL + busy_timeout remain as backstops.
	sqldb.SetMaxOpenConns(1)

	if err := sqldb.PingContext(ctx); err != nil {
		sqldb.Close()
		return nil, err
	}

	db := &DB{sqldb}
	if err := db.migrate(ctx); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// buildDSN appends the connection pragmas to path, accepting either a bare
// filesystem path or an existing "file:" URI with its own query parameters.
func buildDSN(path string) string {
	const pragmas = "_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"

	if !strings.HasPrefix(path, "file:") {
		return "file:" + path + "?" + pragmas
	}
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + pragmas
}
