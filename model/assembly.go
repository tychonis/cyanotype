package model

import (
	"errors"

	"github.com/google/uuid"
)

type Assembly struct {
	ID         uuid.UUID    `json:"id" yaml:"id"`
	Name       string       `json:"name" yaml:"name"`
	PartNumber string       `json:"part_number" yaml:"part_number"`
	Components []*Component `json:"components" yaml:"components"`
}

type Component struct {
	Name string  `json:"name" yaml:"name"`
	Item BOMItem `json:"item" yaml:"item"`
	Qty  float64 `json:"qty" yaml:"qty"`
}

func (a *Assembly) GetID() uuid.UUID {
	return a.ID
}

func (a *Assembly) SetID(id uuid.UUID) error {
	if a.ID != uuid.Nil && a.ID != id {
		return errors.New("id conflict")
	}
	a.ID = id
	return nil
}

func (a *Assembly) GetName() string {
	return a.Name
}

func (a *Assembly) GetPartNumber() string {
	return a.PartNumber
}

func (a *Assembly) GetComponents() []*Component {
	return a.Components
}
