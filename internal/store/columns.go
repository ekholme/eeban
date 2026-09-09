package store

import (
	"context"
	"database/sql"

	"github.com/ekholme/eeban/internal/domain"
)

// CreateColumn inserts a new column at the end of boardID and returns it.
func (db *DB) CreateColumn(ctx context.Context, boardID int64, name string) (domain.Column, error) {
	var maxPos sql.NullInt64
	if err := db.QueryRowContext(ctx,
		`SELECT MAX(position) FROM columns WHERE board_id = ?`, boardID,
	).Scan(&maxPos); err != nil {
		return domain.Column{}, err
	}

	pos := int64(domain.PositionGap)
	if maxPos.Valid {
		pos = maxPos.Int64 + domain.PositionGap
	}

	res, err := db.ExecContext(ctx,
		`INSERT INTO columns (board_id, name, position) VALUES (?, ?, ?)`,
		boardID, name, pos)
	if err != nil {
		return domain.Column{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Column{}, err
	}

	return domain.Column{ID: id, BoardID: boardID, Name: name, Position: pos}, nil
}

// RenameColumn overwrites a column's name.
func (db *DB) RenameColumn(ctx context.Context, id int64, name string) error {
	_, err := db.ExecContext(ctx, `UPDATE columns SET name = ? WHERE id = ?`, name, id)
	return err
}

// SetColumnWIP sets a column's WIP limit, or clears it when limit is nil.
func (db *DB) SetColumnWIP(ctx context.Context, id int64, limit *int) error {
	_, err := db.ExecContext(ctx, `UPDATE columns SET wip_limit = ? WHERE id = ?`, limit, id)
	return err
}

// DeleteColumn permanently removes a column; its cards go with it via
// ON DELETE CASCADE.
func (db *DB) DeleteColumn(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM columns WHERE id = ?`, id)
	return err
}

// MoveColumn relocates columnID to position index toIndex among boardID's
// other columns, renormalizing the board's column positions when the gap
// either side has collapsed.
func (db *DB) MoveColumn(ctx context.Context, boardID, columnID int64, toIndex int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ids, positions, err := boardColumnOrder(ctx, tx, boardID, columnID)
	if err != nil {
		return err
	}

	pos, renormalize := domain.PositionForInsert(positions, toIndex)

	if renormalize {
		if toIndex < 0 {
			toIndex = 0
		}
		if toIndex > len(ids) {
			toIndex = len(ids)
		}
		ordered := make([]int64, 0, len(ids)+1)
		ordered = append(ordered, ids[:toIndex]...)
		ordered = append(ordered, columnID)
		ordered = append(ordered, ids[toIndex:]...)

		newPositions := domain.RenormalizePositions(ordered)
		for id, p := range newPositions {
			if _, err := tx.ExecContext(ctx,
				`UPDATE columns SET position = ? WHERE id = ?`, p, id); err != nil {
				return err
			}
		}
		pos = newPositions[columnID]
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE columns SET position = ? WHERE id = ?`, pos, columnID); err != nil {
		return err
	}

	return tx.Commit()
}

// boardColumnOrder returns the ids and positions of boardID's columns,
// ordered by position, excluding excludeID.
func boardColumnOrder(ctx context.Context, tx *sql.Tx, boardID, excludeID int64) ([]int64, []int64, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, position FROM columns
		  WHERE board_id = ? AND id != ?
		  ORDER BY position`, boardID, excludeID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var ids, positions []int64
	for rows.Next() {
		var id, pos int64
		if err := rows.Scan(&id, &pos); err != nil {
			return nil, nil, err
		}
		ids = append(ids, id)
		positions = append(positions, pos)
	}
	return ids, positions, rows.Err()
}
