// Command eeban runs the TUI kanban board.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/service"
	"github.com/ekholme/eeban/internal/store"
	"github.com/ekholme/eeban/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "eeban:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	dbPath, err := databasePath()
	if err != nil {
		return err
	}

	db, err := store.Open(ctx, dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	svc := service.New(db)

	board, err := svc.DefaultBoard(ctx)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}

	_, err = tea.NewProgram(tui.New(svc, board), tea.WithAltScreen()).Run()
	return err
}

// databasePath honours $EEBAN_DB when set, otherwise the XDG data dir.
func databasePath() (string, error) {
	if p := os.Getenv("EEBAN_DB"); p != "" {
		return p, nil
	}
	return xdg.DataFile("eeban/eeban.db")
}
