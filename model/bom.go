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

type SymbolicRef struct {
	Name   string  `json:"name" yaml:"name"`
	Kind   string  `json:"kind" yaml:"kind"`
	Target BOMItem `json:"target" yaml:"target"`
}

func (r *SymbolicRef) GetID() uuid.UUID {
	if r.Target != nil {
		return r.Target.GetID()
	}
	return uuid.Nil
}

func (r *SymbolicRef) SetID(id uuid.UUID) error {
	return nil
}

func (r *SymbolicRef) GetName() string {
	if r.Target != nil {
		return r.Target.GetName()
	}
	return "UNRESOLVED"
}

func (r *SymbolicRef) GetPartNumber() string {
	if r.Target != nil {
		return r.Target.GetPartNumber()
	}
	return "UNRESOLVED"
}

func (r *SymbolicRef) GetComponents() []*Component {
	if r.Target != nil {
		return r.Target.GetComponents()
	}
	return nil
}

func (r *SymbolicRef) GetDetails() map[string]any {
	return nil
}
