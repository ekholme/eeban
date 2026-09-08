package domain

// PositionForInsert computes the position to assign an item inserted at index
// at among the given ordered positions (which must not include the item being
// inserted/moved). It reports whether the resulting gap has collapsed enough
// that the column should be renormalized.
func PositionForInsert(positions []int64, at int) (pos int64, renormalize bool) {
	if at < 0 {
		at = 0
	}
	if at > len(positions) {
		at = len(positions)
	}

	var before, after *int64
	if at > 0 {
		before = &positions[at-1]
	}
	if at < len(positions) {
		after = &positions[at]
	}

	switch {
	case before == nil && after == nil:
		return PositionGap, false
	case before == nil:
		return *after / 2, *after < 2
	case after == nil:
		return *before + PositionGap, false
	default:
		gap := *after - *before
		return *before + gap/2, gap < 2
	}
}

// RenormalizePositions returns freshly spaced positions (multiples of
// PositionGap) for ids, keeping their given order.
func RenormalizePositions(ids []int64) map[int64]int64 {
	out := make(map[int64]int64, len(ids))
	for i, id := range ids {
		out[id] = int64(i+1) * PositionGap
	}
	return out
}
