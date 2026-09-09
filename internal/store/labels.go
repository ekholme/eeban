package store

import (
	"context"

	"github.com/ekholme/eeban/internal/domain"
)

// CreateLabel inserts a new label on boardID and returns it.
func (db *DB) CreateLabel(ctx context.Context, boardID int64, name, color string) (domain.Label, error) {
	res, err := db.ExecContext(ctx,
		`INSERT INTO labels (board_id, name, color) VALUES (?, ?, ?)`,
		boardID, name, color)
	if err != nil {
		return domain.Label{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return domain.Label{}, err
	}
	return domain.Label{ID: id, BoardID: boardID, Name: name, Color: color}, nil
}

// DeleteLabel permanently removes a label; its card_labels rows go with it via
// ON DELETE CASCADE.
func (db *DB) DeleteLabel(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM labels WHERE id = ?`, id)
	return err
}

// SetCardLabel attaches labelID to cardID when on is true, or detaches it when
// false. Both directions are idempotent.
func (db *DB) SetCardLabel(ctx context.Context, cardID, labelID int64, on bool) error {
	if on {
		_, err := db.ExecContext(ctx,
			`INSERT OR IGNORE INTO card_labels (card_id, label_id) VALUES (?, ?)`,
			cardID, labelID)
		return err
	}
	_, err := db.ExecContext(ctx,
		`DELETE FROM card_labels WHERE card_id = ? AND label_id = ?`, cardID, labelID)
	return err
}
