package store

import (
	"context"

	"github.com/ekholme/eeban/internal/domain"
)

// LoadDueCards returns every non-archived card on any board that has a due
// date, annotated with its board and column, with labels attached. Rows are
// ordered by due date, then board, column, and card position; callers decide
// which dates are relevant (see domain.UpcomingDue).
func (db *DB) LoadDueCards(ctx context.Context) ([]domain.DueCard, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT c.id, c.column_id, c.title, c.body, c.position, c.priority, c.due_date,
		        b.id, b.name, col.name
		   FROM cards c
		   JOIN columns col ON col.id = c.column_id
		   JOIN boards b    ON b.id = col.board_id
		  WHERE c.archived_at IS NULL AND c.due_date IS NOT NULL AND c.due_date != ''
		  ORDER BY c.due_date, b.id, col.position, c.position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []domain.DueCard
	idx := make(map[int64]int)
	for rows.Next() {
		var c domain.DueCard
		if err := rows.Scan(
			&c.ID, &c.ColumnID, &c.Title, &c.Body, &c.Position, &c.Priority, &c.DueDate,
			&c.BoardID, &c.BoardName, &c.ColumnName,
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
		   JOIN cards c  ON c.id = cl.card_id
		  WHERE c.archived_at IS NULL AND c.due_date IS NOT NULL AND c.due_date != ''
		  ORDER BY l.name`)
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
