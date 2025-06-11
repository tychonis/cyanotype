package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

type Item struct {
	ID         uuid.UUID    `json:"id" yaml:"id"`
	Name       string       `json:"name" yaml:"name"`
	Source     string       `json:"source,omitempty" yaml:"source,omitempty"`
	PartNumber string       `json:"part_number" yaml:"part_number"`
	Reference  string       `json:"ref,omitempty" yaml:"ref,omitempty"`
	Components []*Component `json:"-" yaml:"-"`
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

// TODO: implement attrs
func (i *Item) Resolve(path []string) (Symbol, error) {
	return nil, nil
}

func sha256FromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}

	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum), nil
}

func (i *Item) HashReference() string {
	component := strings.Split(i.Reference, ":")
	if len(component) != 2 {
		return ""
	}
	switch component[0] {
	case "file":
		sha, _ := sha256FromFile(component[1])
		return sha
	default:
		slog.Warn("Unsupported scheme", "scheme", component[0])
	}
	return ""
}

func (i *Item) GetDetails() map[string]any {
	details := make(map[string]any)
	details["ref_hash"] = i.HashReference()
	return details
}
