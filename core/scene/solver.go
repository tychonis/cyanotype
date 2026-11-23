package scene

import (
	"fmt"

	"github.com/tychonis/cyanotype/internal/geometry"
	"github.com/tychonis/cyanotype/model"
)

type AlignPoint struct {
	Child string
	Frame model.Placement
}

type AlignConstraint struct {
	A AlignPoint
	B AlignPoint
}

func solveAlignConstraint(
	c AlignConstraint,
	placements map[string]model.Placement,
) (bool, error) {
	a, hasA := placements[c.A.Child]
	b, hasB := placements[c.B.Child]

	switch {
	case hasA && hasB:
		expected := a.Align(c.A.Frame, c.B.Frame)
		if !geometry.SamePlacement(expected, b) {
			return false, fmt.Errorf(
				"conflicting alignment: %s -> %s",
				c.A.Child, c.B.Child,
			)
		}
		return false, nil

	case hasA:
		placements[c.B.Child] = a.Align(c.A.Frame, c.B.Frame)
		return true, nil

	case hasB:
		placements[c.A.Child] = b.Align(c.B.Frame, c.A.Frame)
		return true, nil
	}

	return false, nil
}

func solveAlignConstraints(
	constraints []AlignConstraint,
	placements map[string]model.Placement,
) error {
	for {
		progress := false

		for _, c := range constraints {
			changed, err := solveAlignConstraint(c, placements)
			if err != nil {
				return err
			}
			progress = progress || changed
		}

		if !progress {
			break
		}
	}

	for _, c := range constraints {
		if _, ok := placements[c.A.Child]; !ok {
			return fmt.Errorf("unresolved alignment: %s", c.A.Child)
		}
		if _, ok := placements[c.B.Child]; !ok {
			return fmt.Errorf("unresolved alignment: %s", c.B.Child)
		}
	}

	return nil
}
