package symbols_test

import (
	"testing"

	"github.com/tychonis/cyanotype/internal/symbols"
	"github.com/tychonis/cyanotype/model"
)

func TestRefImplementInterfaces(t *testing.T) {
	var _ model.Symbol = (*symbols.Ref)(nil)
	var _ model.BOMItem = (*symbols.Ref)(nil)
}
