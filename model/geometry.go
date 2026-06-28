package model

type Quaternion [4]float64
type Vec3 [3]float64
type Placement struct {
	Rotation Quaternion `json:"rotation" yaml:"rotation"`
	Position Vec3       `json:"position" yaml:"position"`
}

var IdentityQuaternion = Quaternion{0, 0, 0, 1}
var IdentityVec3 = Vec3{0, 0, 0}

var IdentityPlacement = Placement{
	Rotation: IdentityQuaternion,
	Position: IdentityVec3,
}
