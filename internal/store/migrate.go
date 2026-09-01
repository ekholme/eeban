package store

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// migrate applies every migration file not yet recorded in schema_migrations,
// each inside its own transaction, in ascending numeric order.
func (db *DB) migrate(ctx context.Context) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`); err != nil {
		return err
	}

	applied, err := db.appliedVersions(ctx)
	if err != nil {
		return err
	}

	files, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, f := range files {
		version, err := versionFromName(f)
		if err != nil {
			return err
		}
		if applied[version] {
			continue
		}

		body, err := migrationsFS.ReadFile(f)
		if err != nil {
			return err
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			tx.Rollback()
			return fmt.Errorf("apply %s: %w", f, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record %s: %w", f, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) appliedVersions(ctx context.Context) (map[int]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// versionFromName parses the leading integer of a migration filename:
// "migrations/001_init.sql" -> 1.
func versionFromName(name string) (int, error) {
	base := name[strings.LastIndex(name, "/")+1:]
	prefix, _, found := strings.Cut(base, "_")
	if !found {
		return 0, fmt.Errorf("migration %q missing NNN_ prefix", name)
	}
	v, err := strconv.Atoi(prefix)
	if err != nil {
		return 0, fmt.Errorf("migration %q has non-numeric prefix: %w", name, err)
	}
	return v, nil
}
