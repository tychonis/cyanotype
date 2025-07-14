package distance_test

import (
	"strings"
	"testing"

	"github.com/tychonis/cyanotype/internal/distance"
)

func TestEditDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		wantDist int
	}{
		{"piece.bishop", "piece.bishop", 0},
		{"piece.bishop", "piece.knight", 1},
		{"piece.bishop", "bishop.piece", 2},
		{"piece.bishop", "piece", 1},
		{"piece", "piece.bishop", 1},
		{"pawn", "queen", 1},
		{"rook", "knight", 2},
		{"piece.bishop.extra", "piece.bishop", 1},
		{"a.b.c", "a.b.c.d.e", 2},
	}

	for _, tt := range tests {
		aTokens := strings.Split(tt.a, ".")
		bTokens := strings.Split(tt.b, ".")
		got := distance.EditDistance(aTokens, bTokens)
		if got != tt.wantDist {
			t.Errorf("EditDistance(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.wantDist)
		}
	}
}
