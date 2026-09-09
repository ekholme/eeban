package store

import (
	"context"

	"github.com/ekholme/eeban/internal/domain"
)

// ListBoards returns every board (id and name only), ordered by id.
func (db *DB) ListBoards(ctx context.Context) ([]domain.Board, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, name FROM boards ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		var b domain.Board
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, rows.Err()
}

// CreateBoard inserts a new board seeded with the default three columns and
// returns it (id and name only).
func (db *DB) CreateBoard(ctx context.Context, name string) (domain.Board, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Board{}, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO boards (name) VALUES (?)`, name)
	if err != nil {
		return domain.Board{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Board{}, err
	}

	for i, col := range []string{"Backlog", "In Progress", "Done"} {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO columns (board_id, name, position) VALUES (?, ?, ?)`,
			id, col, int64(i+1)*domain.PositionGap); err != nil {
			return domain.Board{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return domain.Board{}, err
	}
	return domain.Board{ID: id, Name: name}, nil
}

// RenameBoard overwrites a board's name.
func (db *DB) RenameBoard(ctx context.Context, id int64, name string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE boards SET name = ?, updated_at = datetime('now') WHERE id = ?`, name, id)
	return err
}

// DeleteBoard permanently removes a board and everything on it.
func (db *DB) DeleteBoard(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM boards WHERE id = ?`, id)
	return err
}

// LoadBoard returns the board with the given id including its columns and their
// cards, each ordered by position, plus the labels defined for the board.
// Archived cards are excluded.
func (db *DB) LoadBoard(ctx context.Context, id int64) (domain.Board, error) {
	var b domain.Board
	err := db.QueryRowContext(ctx,
		`SELECT id, name FROM boards WHERE id = ?`, id,
	).Scan(&b.ID, &b.Name)
	if err != nil {
		return domain.Board{}, err
	}

	labels, err := db.loadLabels(ctx, id)
	if err != nil {
		return domain.Board{}, err
	}
	b.Labels = labels

	cols, err := db.loadColumns(ctx, id)
	if err != nil {
		return domain.Board{}, err
	}
	b.Columns = cols
	return b, nil
}

func (db *DB) loadColumns(ctx context.Context, boardID int64) ([]domain.Column, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, board_id, name, position, wip_limit
		   FROM columns
		  WHERE board_id = ?
		  ORDER BY position`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []domain.Column
	indexByID := make(map[int64]int)
	for rows.Next() {
		var c domain.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Name, &c.Position, &c.WIPLimit); err != nil {
			return nil, err
		}
		indexByID[c.ID] = len(cols)
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return cols, nil
	}

	crows, err := db.QueryContext(ctx,
		`SELECT c.id, c.column_id, c.title, c.body, c.position, c.priority, c.due_date
		   FROM cards c
		   JOIN columns col ON col.id = c.column_id
		  WHERE col.board_id = ? AND c.archived_at IS NULL
		  ORDER BY c.position`, boardID)
	if err != nil {
		return nil, err
	}
	defer crows.Close()

	// cardLoc maps a card id to its (column index, card index) so a later
	// labels pass can attach rows without re-walking every column.
	cardLoc := make(map[int64][2]int)
	for crows.Next() {
		var c domain.Card
		if err := crows.Scan(
			&c.ID, &c.ColumnID, &c.Title, &c.Body, &c.Position, &c.Priority, &c.DueDate,
		); err != nil {
			return nil, err
		}
		if idx, ok := indexByID[c.ColumnID]; ok {
			cardLoc[c.ID] = [2]int{idx, len(cols[idx].Cards)}
			cols[idx].Cards = append(cols[idx].Cards, c)
		}
	}
	if err := crows.Err(); err != nil {
		return nil, err
	}

	lrows, err := db.QueryContext(ctx,
		`SELECT cl.card_id, l.id, l.board_id, l.name, l.color
		   FROM card_labels cl
		   JOIN labels l ON l.id = cl.label_id
		  WHERE l.board_id = ?
		  ORDER BY l.name`, boardID)
	if err != nil {
		return nil, err
	}
	defer lrows.Close()

	for lrows.Next() {
		var cardID int64
		var l domain.Label
		if err := lrows.Scan(&cardID, &l.ID, &l.BoardID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		if loc, ok := cardLoc[cardID]; ok {
			cols[loc[0]].Cards[loc[1]].Labels = append(cols[loc[0]].Cards[loc[1]].Labels, l)
		}
	}
	return cols, lrows.Err()
}

// loadLabels returns every label defined for boardID, ordered by name.
func (db *DB) loadLabels(ctx context.Context, boardID int64) ([]domain.Label, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT id, board_id, name, color FROM labels WHERE board_id = ? ORDER BY name`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var labels []domain.Label
	for rows.Next() {
		var l domain.Label
		if err := rows.Scan(&l.ID, &l.BoardID, &l.Name, &l.Color); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}
