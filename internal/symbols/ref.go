package symbols

import (
	"errors"

	"github.com/google/uuid"
	"github.com/tychonis/cyanotype/model"
)

type Ref struct {
	Name   string
	Kind   string
	Target model.BOMItem
}

func (r *Ref) Resolve(path []string) (model.Symbol, error) {
	t, ok := r.Target.(model.Symbol)
	if !ok {
		return nil, errors.New("resolution failed")
	}
	return t.Resolve(path)
}

func (r *Ref) GetID() uuid.UUID {
	if r.Target != nil {
		return r.Target.GetID()
	}
	return uuid.Nil
}

func (r *Ref) SetID(id uuid.UUID) error {
	return nil
}

func (r *Ref) GetName() string {
	if r.Target != nil {
		return r.Target.GetName()
	}
	return "UNRESOLVED"
}

func (r *Ref) GetPartNumber() string {
	if r.Target != nil {
		return r.Target.GetPartNumber()
	}
	return "UNRESOLVED"
}

func (r *Ref) GetComponents() []*model.Component {
	if r.Target != nil {
		return r.Target.GetComponents()
	}
	return nil
}

func (r *Ref) GetDetails() map[string]any {
	return nil
}
