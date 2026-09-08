package domain

import (
	"reflect"
	"testing"
)

func TestPositionForInsert(t *testing.T) {
	cases := []struct {
		name       string
		positions  []int64
		at         int
		wantPos    int64
		wantRenorm bool
	}{
		{"empty column", nil, 0, PositionGap, false},
		{"insert at start", []int64{1000, 2000}, 0, 500, false},
		{"insert in middle", []int64{1000, 2000}, 1, 1500, false},
		{"insert at end", []int64{1000, 2000}, 2, 3000, false},
		{"index clamped past end", []int64{1000, 2000}, 5, 3000, false},
		{"index clamped below zero", []int64{1000, 2000}, -1, 500, false},
		{"collapsed gap at start needs renorm", []int64{1, 2000}, 0, 0, true},
		{"collapsed gap in middle needs renorm", []int64{1000, 1001}, 1, 1000, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pos, renorm := PositionForInsert(tc.positions, tc.at)
			if pos != tc.wantPos || renorm != tc.wantRenorm {
				t.Errorf("PositionForInsert(%v, %d) = (%d, %v), want (%d, %v)",
					tc.positions, tc.at, pos, renorm, tc.wantPos, tc.wantRenorm)
			}
		})
	}
}

func TestRenormalizePositions(t *testing.T) {
	got := RenormalizePositions([]int64{7, 3, 9})
	want := map[int64]int64{7: 1000, 3: 2000, 9: 3000}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("RenormalizePositions = %v, want %v", got, want)
	}

	if got := RenormalizePositions(nil); len(got) != 0 {
		t.Errorf("RenormalizePositions(nil) = %v, want empty", got)
	}
}
