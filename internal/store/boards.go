package store

import (
	"context"

	"github.com/ekholme/eeban/internal/domain"
)

// LoadBoard returns the board with the given id including its columns and their
// cards, each ordered by position. Archived cards are excluded.
func (db *DB) LoadBoard(ctx context.Context, id int64) (domain.Board, error) {
	var b domain.Board
	err := db.QueryRowContext(ctx,
		`SELECT id, name FROM boards WHERE id = ?`, id,
	).Scan(&b.ID, &b.Name)
	if err != nil {
		return domain.Board{}, err
	}

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

	for crows.Next() {
		var c domain.Card
		if err := crows.Scan(
			&c.ID, &c.ColumnID, &c.Title, &c.Body, &c.Position, &c.Priority, &c.DueDate,
		); err != nil {
			return nil, err
		}
		if idx, ok := indexByID[c.ColumnID]; ok {
			cols[idx].Cards = append(cols[idx].Cards, c)
		}
	}
	return cols, crows.Err()
}
