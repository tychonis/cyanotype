package model

import "github.com/google/uuid"

type BOMItem interface {
	GetID() uuid.UUID
	SetID(id uuid.UUID) error
	GetName() string
	GetPartNumber() string
	GetComponents() []*Component
	GetDetails() map[string]any
}
