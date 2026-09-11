package model_test

import (
	"math"
	"testing"

	"github.com/tychonis/cyanotype/model"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestPlacementComposeTranslation(t *testing.T) {
	childToParent := model.Placement{
		Rotation: model.IdentityQuaternion,
		Position: model.Vec3{1, 2, 3},
	}
	parentToGrandparent := model.Placement{
		Rotation: model.IdentityQuaternion,
		Position: model.Vec3{10, 20, 30},
	}

	got := childToParent.Compose(&parentToGrandparent)

	want := model.Vec3{11, 22, 33}
	for i := range 3 {
		if !almostEqual(got.Position[i], want[i]) {
			t.Fatalf("position = %v, want %v", got.Position, want)
		}
	}
}

func TestPlacementComposeRotation(t *testing.T) {
	// Parent is rotated 90° around Z.
	s := math.Sqrt(0.5)

	childToParent := model.Placement{
		Rotation: model.IdentityQuaternion,
		Position: model.Vec3{1, 0, 0},
	}
	parentToGrandparent := model.Placement{
		Rotation: model.Quaternion{0, 0, s, s},
		Position: model.IdentityVec3,
	}

	got := childToParent.Compose(&parentToGrandparent)

	want := model.Vec3{0, 1, 0}
	for i := range 3 {
		if !almostEqual(got.Position[i], want[i]) {
			t.Fatalf("position = %v, want %v", got.Position, want)
		}
	}
}

func TestPlacementComposeIdentity(t *testing.T) {
	p := model.Placement{
		Rotation: model.Quaternion{0, 0, 0, 1},
		Position: model.Vec3{1, 2, 3},
	}

	got := p.Compose(&model.IdentityPlacement)

	if got.Rotation != p.Rotation || got.Position != p.Position {
		t.Fatalf("got %+v, want %+v", got, p)
	}
}

func TestPlacementComposeRotationAndTranslation(t *testing.T) {
	s := math.Sqrt(0.5)

	// Child: rotate 90° around Z, then translate +X in parent frame.
	childToParent := model.Placement{
		Rotation: model.Quaternion{0, 0, s, s},
		Position: model.Vec3{1, 0, 0},
	}

	// Parent: rotate 90° around Z, then translate +Y in grandparent frame.
	parentToGrandparent := model.Placement{
		Rotation: model.Quaternion{0, 0, s, s},
		Position: model.Vec3{0, 2, 0},
	}

	got := childToParent.Compose(&parentToGrandparent)

	// Rotation: 90° + 90° = 180° around Z.
	wantRotation := model.Quaternion{0, 0, 1, 0}

	// Child translation {1,0,0} is rotated by parent's 90° -> {0,1,0},
	// then parent's translation {0,2,0} is added -> {0,3,0}.
	wantPosition := model.Vec3{0, 3, 0}

	for i := range 4 {
		if !almostEqual(got.Rotation[i], wantRotation[i]) {
			t.Fatalf("rotation = %v, want %v", got.Rotation, wantRotation)
		}
	}

	for i := range 3 {
		if !almostEqual(got.Position[i], wantPosition[i]) {
			t.Fatalf("position = %v, want %v", got.Position, wantPosition)
		}
	}
}
