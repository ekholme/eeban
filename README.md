# eeban

A single-user TUI kanban board. Local-first, keyboard-driven, one SQLite file.

## Status

Early scaffold. Board rendering, navigation, and card create/move/reorder/
delete are wired up; editing a card's body/priority/due date and column
management are not yet. See [PLAN.md](PLAN.md).

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
| `esc` | close the detail pane / cancel new-card input |
| `n` | new card (title only) in the current column |
| `d` | delete the selected card |
| `H` / `L` | move the selected card to the previous / next column |
| `J` / `K` | move the selected card down / up within its column |
| `q` / `ctrl+c` | quit |

## License

[MIT](LICENSE)
