# eeban — Project Plan

A single-user TUI kanban board. Local-first, keyboard-driven, one SQLite file.

## Locked decisions

| Area | Decision |
|---|---|
| Language | Go 1.27 |
| TUI | Bubble Tea + Bubbles + Lipgloss (charmbracelet) |
| Storage | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| Data layer | Hand-written `database/sql`; **no** sqlc, **no** goose |
| Migrations | Embedded `.sql` files + a ~40-line runner over a `schema_migrations` table |
| Boards | `boards` table + `board_id` FKs from day one; MVP UI renders one board (id 1, auto-seeded) |
| Paths (XDG) | DB: `~/.local/share/eeban/eeban.db` · config: `~/.config/eeban/config.toml` (later) |
| Module path | `github.com/ekholme/eeban` (change if your GitHub handle differs) |

Non-goals: multi-user, web UI, real-time sync, auth, a server.

Feature non-goals (keeps the app minimalist as it grows):
- No background daemon or desktop notifications — due-date highlighting already
  covers this in-app; a notifier means a long-running process, which cuts
  against "one SQLite file, run when you run it."
- No interactive checklists — markdown rendering already displays `- [ ]` as
  plain read-only text; making checkboxes clickable means parsing and
  rewriting card body text for little gain.
- No swimlanes, sub-boards, or nested cards — keep the `boards → columns →
  cards` model exactly as flat as it is.
- No custom/arbitrary card fields — title/body/priority/due/labels is the
  whole schema, on purpose.

## Layout

```
cmd/eeban/main.go            entrypoint: XDG path, open store, load board, run TUI
internal/
  domain/                    Board / Column / Card types + rules (move, reorder, WIP)
  store/
    db.go                    Open(): *sql.DB, pragmas, MaxOpenConns(1), run migrations
    migrate.go               embedded migration runner
    migrations/001_init.sql  schema + seed board
    boards.go                board reads (LoadBoard = board + columns + cards)
    columns.go               column CRUD / reorder ✅
    cards.go                 card create/update/delete/move ✅
    labels.go                label CRUD + card_labels join ✅
  service/                   the only API the TUI calls; orchestrates domain + store
  tui/
    model.go                 root Bubble Tea model, Update loop, cursor state
    keys.go                  KeyMap (vim + arrows)
    styles.go                Lipgloss styles (adaptive light/dark)
    view.go                  board / column / card rendering, layout math
    helpers.go               small pure helpers
```

Rule: the TUI never touches SQL. It calls `service`, gets state back, re-renders.
DB calls that mutate get wrapped in `tea.Cmd` so the UI never blocks.

## Data model (001_init.sql)

```
boards       (id, name, created_at, updated_at)
columns      (id, board_id→boards, name, position, wip_limit NULL)
cards        (id, column_id→columns, title, body, position, priority,
              due_date NULL, created_at, updated_at, archived_at NULL)
```

Later migrations add: `labels`, `card_labels` ✅ (`002_labels.sql`). `card_events`
is not yet migrated — it's a prerequisite for the Stats item under Later, not
something to add speculatively ahead of it.

**Ordering:** integer `position` with gaps of 1000 (1000, 2000, 3000…). A move
rewrites one row's position to the midpoint of its neighbours; renormalize a
column to clean multiples of 1000 when gaps get too small (< 2).

**SQLite setup:** DSN pragmas `busy_timeout(5000)`, `journal_mode(WAL)`,
`foreign_keys(ON)`; `db.SetMaxOpenConns(1)` so all access is serialized and
writes never hit `SQLITE_BUSY`. WAL + busy_timeout stay as backstops.

## Feature phases

### MVP
- [x] Repo scaffold, migration runner, seeded board
- [x] Read-only board render: columns side by side, cards, selection cursor
- [x] Navigate cards/columns (`h/j/k/l` + arrow keys), resize handling
- [x] Card detail pane for the selected card (`enter` toggles, `esc` closes)
- [x] Create card (title only, `n`) / edit body+priority+due date (`e`) / delete card (`d`)
- [x] Move card across columns (`H`/`L`); reorder within a column (`J`/`K`)
- [x] Add / rename / delete / reorder columns (`N`/`R`/`D`/`[`/`]`)
- [x] Help overlay (`?`), confirm dialog, transient error toast
- [x] Every mutation persisted immediately

### v1
- [x] Labels with colors; filter by label (`t` picker, `f` filters by the highlighted label)
- [x] Fuzzy search / filter across title + body (`/`, live subsequence match)
- [x] Due-date highlighting (overdue / due soon) — `domain.DueStatusFor`
- [x] WIP limits with warning styling (`w` sets a column limit; over-limit columns turn amber)
- [x] Archive view for done cards (`a` archives, `A` opens the list, `r` restores)
- [x] Multiple boards + board switcher (`b`: switch / create / rename / delete)
- [x] Single-step undo (`u`) — in-memory, covers create / move / reorder / archive

### Later
- [ ] `$EDITOR` integration for card bodies
- [x] Markdown rendering in detail pane (glamour)
- [ ] Duplicate/clone card (`c`) — same column, appended below the source,
  copies title/body/priority/labels; small addition on top of `CreateCard`
- [ ] Multi-step undo — bump the existing single-step undo to a small bounded
  in-memory stack (~20 actions); still no persistence, just a deeper history
- [ ] JSON / Markdown import-export
- [ ] Themes + configurable keymap (config.toml)
- [ ] Backup/sync: Litestream, or git the DB file
- [ ] `card_events` migration, then stats: throughput / cycle time

### v2 (maybe)
Bigger, optional bets — split out from Later on purpose so they need
deliberate buy-in before starting, not just idle time.
- [ ] Global search across boards — `/` today only filters the current board;
  searching across all boards needs a new service method and a cross-board
  result list
- [ ] Bulk/multi-select actions — visual-mode-style card selection to move,
  archive, or label several cards at once; real complexity (selection state,
  multi-cursor rendering), not a given

## Build sequence

1. **Scaffold** — go.mod, justfile, migration runner, DB open with pragmas, seed. ✅
2. **Read model** — `store.LoadBoard` returns board + columns + cards ordered by position. ✅
3. **Read-only TUI** — render seeded board, navigation, `WindowSizeMsg` layout. ✅
4. **Domain + service mutators** — create/move/reorder card, with unit tests against `file::memory:`. ✅
5. **Wire mutations into the TUI** — card CRUD, move/reorder, as `tea.Cmd`s that reload the board. ✅
6. **Column management.** ✅
7. **Card form + detail pane** — Bubbles `textinput` / `textarea`, nested model + focus. ✅ (detail pane, title quick-add, full body/priority/due-date edit form via `e`)
8. **Help, confirms, toasts, styling pass.** ✅
9. **Iterate on v1.** ✅ (labels, search, due dates, WIP, archive, multi-board, undo)

## Risks / decisions to nail early

- **Layout math** — N columns wider than the terminal: clamp column width, then
  add horizontal scroll of columns once `sum(width) > terminal width`.
- **Nested model / focus** — form and modal state in Bubble Tea; pick a pattern
  (sub-model returns `(sub, cmd)`, root delegates while modal is active) in step 7.
- **Reorder strategy** — implement the gap/midpoint + renormalize helper once in
  `domain` and reuse for cards and columns.
- **TUI tests** — `teatest` golden files for the key interaction flows.

## Testing

- `store` / `service`: open `file::memory:?cache=shared`, run migrations, exercise repos.
- `domain`: pure unit tests for move/reorder/renormalize.
- `tui`: `teatest` for navigation and form flows.

## Commands (justfile)

- `just run` — build and run against the real XDG database
- `just dev` — run against a throwaway `./eeban.dev.db`
- `just test` — `go test ./...`
- `just build` — static binary to `./bin/eeban`
