package model

import "github.com/google/uuid"

type TransformationID = uuid.UUID

type Transformation struct {
	ID TransformationID
}
