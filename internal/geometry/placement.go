package geometry

import (
	"math"

	"github.com/tychonis/cyanotype/model"
)

func NormalizeQuaternion(q model.Quaternion) model.Quaternion {
	n := math.Sqrt(
		q[0]*q[0] + q[1]*q[1] +
			q[2]*q[2] + q[3]*q[3],
	)
	if n == 0 {
		return model.IdentityQuaternion
	}
	return model.Quaternion{
		q[0] / n,
		q[1] / n,
		q[2] / n,
		q[3] / n,
	}
}

func SamePlacement(a, b model.Placement) bool {
	const posTol = 1e-6 // Assuming mm.
	const rotTol = 1e-8 // Radians.

	dx := a.Position[0] - b.Position[0]
	dy := a.Position[1] - b.Position[1]
	dz := a.Position[2] - b.Position[2]

	if math.Sqrt(dx*dx+dy*dy+dz*dz) > posTol {
		return false
	}

	// Normalize first to avoid scaling errors.
	qa := NormalizeQuaternion(a.Rotation)
	qb := NormalizeQuaternion(b.Rotation)

	dot := math.Abs(
		qa[0]*qb[0] + qa[1]*qb[1] +
			qa[2]*qb[2] + qa[3]*qb[3],
	)
	dot = math.Min(1, dot)

	return 2*math.Acos(dot) <= rotTol
}
