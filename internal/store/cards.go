package store

import (
	"context"
	"database/sql"

	"github.com/ekholme/eeban/internal/domain"
)

// CreateCard inserts a new card at the end of columnID and returns it.
func (db *DB) CreateCard(ctx context.Context, columnID int64, title string) (domain.Card, error) {
	var maxPos sql.NullInt64
	if err := db.QueryRowContext(ctx,
		`SELECT MAX(position) FROM cards WHERE column_id = ? AND archived_at IS NULL`, columnID,
	).Scan(&maxPos); err != nil {
		return domain.Card{}, err
	}

	pos := int64(domain.PositionGap)
	if maxPos.Valid {
		pos = maxPos.Int64 + domain.PositionGap
	}

	res, err := db.ExecContext(ctx,
		`INSERT INTO cards (column_id, title, position) VALUES (?, ?, ?)`,
		columnID, title, pos)
	if err != nil {
		return domain.Card{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Card{}, err
	}

	return domain.Card{ID: id, ColumnID: columnID, Title: title, Position: pos}, nil
}

// UpdateCard overwrites the editable fields of an existing card.
func (db *DB) UpdateCard(ctx context.Context, id int64, title, body string, priority int, dueDate *string) error {
	_, err := db.ExecContext(ctx,
		`UPDATE cards
		    SET title = ?, body = ?, priority = ?, due_date = ?, updated_at = datetime('now')
		  WHERE id = ?`,
		title, body, priority, dueDate, id)
	return err
}

// DeleteCard permanently removes a card.
func (db *DB) DeleteCard(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM cards WHERE id = ?`, id)
	return err
}

// MoveCard relocates cardID into toColumnID at position index toIndex among
// that column's other (non-archived) cards, renormalizing the column's
// positions when the gap either side has collapsed. Moving within the same
// column reorders it.
func (db *DB) MoveCard(ctx context.Context, cardID, toColumnID int64, toIndex int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ids, positions, err := columnCardOrder(ctx, tx, toColumnID, cardID)
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
		ordered = append(ordered, cardID)
		ordered = append(ordered, ids[toIndex:]...)

		newPositions := domain.RenormalizePositions(ordered)
		for id, p := range newPositions {
			if _, err := tx.ExecContext(ctx,
				`UPDATE cards SET position = ? WHERE id = ?`, p, id); err != nil {
				return err
			}
		}
		pos = newPositions[cardID]
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE cards SET column_id = ?, position = ?, updated_at = datetime('now') WHERE id = ?`,
		toColumnID, pos, cardID); err != nil {
		return err
	}

	return tx.Commit()
}

// columnCardOrder returns the ids and positions of columnID's non-archived
// cards, ordered by position, excluding excludeID.
func columnCardOrder(ctx context.Context, tx *sql.Tx, columnID, excludeID int64) ([]int64, []int64, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, position FROM cards
		  WHERE column_id = ? AND archived_at IS NULL AND id != ?
		  ORDER BY position`, columnID, excludeID)
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
