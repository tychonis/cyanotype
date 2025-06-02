package model

import (
	"errors"

	"github.com/google/uuid"
)

type Part struct {
	ID         uuid.UUID `json:"id" yaml:"id"`
	Name       string    `json:"name" yaml:"name"`
	PartNumber string    `json:"part_number" yaml:"part_number"`
	Reference  string    `json:"ref" yaml:"ref"`
}

func (p *Part) GetID() uuid.UUID {
	return p.ID
}

func (p *Part) SetID(id uuid.UUID) error {
	if p.ID != uuid.Nil && p.ID != id {
		return errors.New("id conflict")
	}
	p.ID = id
	return nil
}

func (p *Part) GetName() string {
	return p.Name
}

func (p *Part) GetPartNumber() string {
	return p.PartNumber
}

func (p *Part) GetComponents() []*Component {
	return nil
}
