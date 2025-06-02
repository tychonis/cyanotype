package model

import (
	"errors"

	"github.com/google/uuid"
)

type Item struct {
	ID         uuid.UUID    `json:"id" yaml:"id"`
	Name       string       `json:"name" yaml:"name"`
	Source     string       `json:"source" yaml:"source"`
	PartNumber string       `json:"part_number" yaml:"part_number"`
	Reference  string       `json:"ref" yaml:"ref"`
	Components []*Component `json:"components" yaml:"components"`
}

type Component struct {
	Name string  `json:"name" yaml:"name"`
	Item BOMItem `json:"item" yaml:"item"`
	Qty  float64 `json:"qty" yaml:"qty"`
}

func (i *Item) GetID() uuid.UUID {
	return i.ID
}

func (i *Item) SetID(id uuid.UUID) error {
	if i.ID != uuid.Nil && i.ID != id {
		return errors.New("id conflict")
	}
	i.ID = id
	return nil
}

func (i *Item) GetName() string {
	return i.Name
}

func (i *Item) GetPartNumber() string {
	return i.PartNumber
}

func (i *Item) GetComponents() []*Component {
	return i.Components
}
