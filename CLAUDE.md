# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`eeban` is a single-user, local-first TUI kanban board: Go + Bubble Tea, one SQLite
file, keyboard-driven. The MVP and the v1 feature set are implemented: full card
and column CRUD/move/reorder, labels, fuzzy search, due-date highlighting, WIP
limits, an archive view, multiple boards, and single-step undo. `PLAN.md` holds
the locked decisions, data model, feature phases, and build sequence; consult it
before adding features.

## Commands

```sh
just run          # run against the real XDG db (~/.local/share/eeban/eeban.db)
just dev          # run against a throwaway ./eeban.dev.db (sets EEBAN_DB)
just test         # go test ./...
just build        # static binary (CGO_ENABLED=0) to ./bin/eeban
just tidy         # go mod tidy

go test ./internal/store -run TestOpenRunsMigrationsAndSeeds   # single test
go vet ./...
```

`EEBAN_DB` overrides the database path (a plain path or a full `file:` DSN); tests
rely on this.

## Architecture

Strict one-way layering — each layer only imports the ones below it:

```
cmd/eeban/main.go   XDG path → store.Open → service.New → tui.New → tea.NewProgram
internal/tui        Bubble Tea front end; calls service only, NEVER touches SQL
internal/service    the single API the TUI uses; orchestrates domain + store
internal/store      hand-written database/sql against SQLite; not imported by tui
internal/domain     pure kanban types + rules; no deps on store or tui
```

The TUI-never-touches-SQL rule is load-bearing. Per `PLAN.md`, mutations get
wrapped in `tea.Cmd` so the UI never blocks on the DB, and the model reloads
board state after a successful write.

### Store & SQLite

- Pure-Go driver `modernc.org/sqlite` (no CGO). `store.Open` sets
  `SetMaxOpenConns(1)` so all access is serialized and writes never hit
  `SQLITE_BUSY`; WAL + `busy_timeout` are backstops. DSN pragmas
  (`busy_timeout`, `journal_mode(WAL)`, `foreign_keys(ON)`) are appended in
  `buildDSN`.
- `DB.LoadBoard` returns a whole `domain.Board` (columns + non-archived cards,
  each ordered by `position`) in two queries.

### Migrations

`internal/store/migrations/NNN_name.sql`, embedded via `//go:embed`. The ~40-line
runner in `migrate.go` applies each unapplied file in its own transaction in
ascending numeric order, tracked in `schema_migrations`. Seed data (the default
board id 1, its columns, sample cards) lives in `001_init.sql`.

**To change the schema, add a new migration file — never edit one that may have
been applied.** Migrations must stay idempotent-at-the-set-level (see
`TestMigrateIsIdempotent`); use `INSERT` seeds only in the first migration.

### Ordering

Items carry an integer `position` spaced by `domain.PositionGap` (1000). A move
rewrites one row to the midpoint of its neighbours; renormalize a column to clean
multiples of 1000 when a gap drops below 2. This helper should live once in
`domain` and be reused for both cards and columns.

## Testing

- `store` / `service`: real SQLite in `t.TempDir()` or `file::memory:?cache=shared`,
  run migrations, exercise the repos.
- `domain`: pure unit tests for move / reorder / renormalize.
- `tui`: drive `Update` / `View` directly with a `nil` service where no DB is
  needed (see `internal/tui/model_test.go`); `teatest` golden files for full
  interaction flows are planned.
