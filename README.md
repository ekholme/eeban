# eeban

A single-user TUI kanban board. Local-first, keyboard-driven, one SQLite file.

## Status

Early scaffold. Read-only board rendering with keyboard navigation works;
mutations are not wired up yet. See [PLAN.md](PLAN.md).

## Run

```sh
just run      # against ~/.local/share/eeban/eeban.db
just dev      # against ./eeban.dev.db (throwaway)
just test
```

## Keys

| Key | Action |
|---|---|
| `h` `l` / `←` `→` | previous / next column |
| `j` `k` / `↓` `↑` | next / previous card |
| `enter` | toggle detail pane for the selected card |
| `esc` | close the detail pane |
| `q` / `ctrl+c` | quit |

## License

[MIT](LICENSE)
