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

type ItemID = uuid.UUID
type NodeID = uuid.UUID

type Item struct {
	ID         ItemID       `json:"id" yaml:"id"`
	Qualifier  string       `json:"qualifier" yaml:"qualifier"`
	Name       string       `json:"name" yaml:"name"`
	Source     string       `json:"source,omitempty" yaml:"source,omitempty"`
	PartNumber string       `json:"part_number" yaml:"part_number"`
	Reference  string       `json:"ref,omitempty" yaml:"ref,omitempty"`
	From       []*Component `json:"-" yaml:"-"`
}

type ItemNode struct {
	ID       NodeID   `json:"id" yaml:"id"`
	Path     string   `json:"path" yaml:"path"`
	ItemID   ItemID   `json:"item_id" yaml:"item_id"`
	ParentID NodeID   `json:"parent_id" yaml:"parent_id"`
	Children []NodeID `json:"children" yaml:"children"`
	Qty      float64  `json:"qty" yaml:"qty"`
}

type Component struct {
	Name string   `json:"name" yaml:"name"`
	Ref  []string `json:"ref" yaml:"ref"`
	Qty  float64  `json:"qty" yaml:"qty"`
}

func (i *Item) GetID() ItemID {
	return i.ID
}

func (i *Item) SetID(id ItemID) error {
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
	return i.From
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

// TODO: implement attrs?
func (i *Item) Resolve(path []string) (Symbol, error) {
	if len(path) > 0 {
		return nil, errors.New("attr not implemented")
	}
	return i, nil
}
