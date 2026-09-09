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

// ArchiveCard marks a card archived, removing it from the board view while
// keeping its row (and label links) intact.
func (db *DB) ArchiveCard(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE cards SET archived_at = datetime('now'), updated_at = datetime('now')
		  WHERE id = ? AND archived_at IS NULL`, id)
	return err
}

// UnarchiveCard clears a card's archived flag and drops it at the end of its
// original column.
func (db *DB) UnarchiveCard(ctx context.Context, id int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var columnID int64
	if err := tx.QueryRowContext(ctx,
		`SELECT column_id FROM cards WHERE id = ?`, id).Scan(&columnID); err != nil {
		return err
	}

	var maxPos sql.NullInt64
	if err := tx.QueryRowContext(ctx,
		`SELECT MAX(position) FROM cards WHERE column_id = ? AND archived_at IS NULL`, columnID,
	).Scan(&maxPos); err != nil {
		return err
	}
	pos := int64(domain.PositionGap)
	if maxPos.Valid {
		pos = maxPos.Int64 + domain.PositionGap
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE cards SET archived_at = NULL, position = ?, updated_at = datetime('now')
		  WHERE id = ?`, pos, id); err != nil {
		return err
	}
	return tx.Commit()
}

// LoadArchived returns boardID's archived cards, most recently archived first,
// with their labels attached.
func (db *DB) LoadArchived(ctx context.Context, boardID int64) ([]domain.Card, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT c.id, c.column_id, c.title, c.body, c.position, c.priority, c.due_date, c.archived_at
		   FROM cards c
		   JOIN columns col ON col.id = c.column_id
		  WHERE col.board_id = ? AND c.archived_at IS NOT NULL
		  ORDER BY c.archived_at DESC, c.id DESC`, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []domain.Card
	idx := make(map[int64]int)
	for rows.Next() {
		var c domain.Card
		if err := rows.Scan(
			&c.ID, &c.ColumnID, &c.Title, &c.Body, &c.Position, &c.Priority, &c.DueDate, &c.ArchivedAt,
		); err != nil {
			return nil, err
		}
		idx[c.ID] = len(cards)
		cards = append(cards, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(cards) == 0 {
		return cards, nil
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
		if i, ok := idx[cardID]; ok {
			cards[i].Labels = append(cards[i].Labels, l)
		}
	}
	return cards, lrows.Err()
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
