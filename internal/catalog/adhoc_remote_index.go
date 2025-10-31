package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tychonis/cyanotype/model"
)

// Adhoc implementation for early prototyping.
// TODO: implement incremental updates. See git negotiation algorithm
type RemoteIndex struct {
	QualifierIndex map[Qualifier]model.ItemID          `json:"qualifier_index"`
	ProcessIndex   map[model.ItemID]*ProcessIndexEntry `json:"process_index"`
	TypeIndex      map[model.ItemID]string             `json:"type_index"`

	Endpoint string `json:"-"`
}

func NewRemoteIndex(endpoint string) *RemoteIndex {
	resp, err := http.Get(endpoint)
	if err != nil {
		return initRemoteIndex(endpoint)
	}
	defer resp.Body.Close()

	var idx RemoteIndex
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&idx)
	if err != nil {
		return initRemoteIndex(endpoint)
	}
	return &idx
}

func initRemoteIndex(endpoint string) *RemoteIndex {
	return &RemoteIndex{
		QualifierIndex: make(map[Qualifier]model.Digest),
		ProcessIndex:   make(map[Qualifier]*ProcessIndexEntry),
		TypeIndex:      make(map[model.Digest]string),

		Endpoint: endpoint,
	}
}

func (idx *RemoteIndex) Save() error {
	content, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	resp, err := http.Post(idx.Endpoint, "application/json", bytes.NewReader(content))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("failed to save index")
	}
	return nil
}

func (idx *RemoteIndex) addToMainIndex(key string, val string) error {
	oldVal, ok := idx.QualifierIndex[key]
	if ok {
		if oldVal == val {
			return nil
		}
	}
	idx.QualifierIndex[key] = val

	return nil
}

func (idx *RemoteIndex) addToTypeIndex(key string, val string) error {
	oldVal, ok := idx.TypeIndex[key]
	if ok {
		if oldVal == val {
			return nil
		}
	}
	idx.TypeIndex[key] = val

	return nil
}

func (idx *RemoteIndex) addToProcessIndex(pType string, key string, val string) error {
	entry, ok := idx.ProcessIndex[key]
	if !ok || entry == nil {
		entry = NewProcessIndexEntry()
		idx.ProcessIndex[key] = entry
	}

	switch pType {
	case "process":
		for _, p := range entry.Processes {
			if p == val {
				return nil
			}
		}
		entry.Processes = append(entry.Processes, val)
	case "coprocess":
		for _, p := range entry.CoProcesses {
			if p == val {
				return nil
			}
		}
		entry.CoProcesses = append(entry.CoProcesses, val)
	default:
		return errors.New("illegal process type")
	}
	return nil
}

func (idx *RemoteIndex) indexProcess(sym model.ConcreteSymbol) error {
	switch resolved := sym.(type) {
	case *model.Process:
		for _, bomLine := range resolved.Input {
			idx.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output {
			idx.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
	case *model.CoProcess:
		for _, bomLine := range resolved.Input {
			idx.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output {
			idx.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
	}
	return nil
}

// TODO: fix this hack.
func (idx *RemoteIndex) indexType(sym model.ConcreteSymbol) error {
	switch sym.(type) {
	case *model.Process:
		idx.addToTypeIndex(sym.GetDigest(), "process")
	case *model.CoProcess:
		idx.addToTypeIndex(sym.GetDigest(), "coprocess")
	case *model.Item:
		idx.addToTypeIndex(sym.GetDigest(), "item")
	case *model.CoItem:
		idx.addToTypeIndex(sym.GetDigest(), "coitem")
	}
	return nil
}

func (idx *RemoteIndex) Index(sym model.ConcreteSymbol) error {
	err := idx.addToMainIndex(sym.GetQualifier(), sym.GetDigest())
	if err != nil {
		return err
	}
	err = idx.indexProcess(sym)
	if err != nil {
		return err
	}
	return idx.indexType(sym)
}

func (idx *RemoteIndex) Find(q Qualifier) (model.Digest, error) {
	digest, ok := idx.QualifierIndex[q]
	if !ok {
		return "", ErrNotFound
	}
	return digest, nil
}

func (idx *RemoteIndex) GetType(digest model.Digest) (string, error) {
	t, ok := idx.TypeIndex[digest]
	if !ok {
		return "", ErrNotFound
	}
	return t, nil
}

func (idx *RemoteIndex) GetItemProcesses(item model.ItemID) ([]model.ProcessID, error) {
	entry, ok := idx.ProcessIndex[item]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.Processes, nil
}

func (idx *RemoteIndex) GetItemCoProcesses(item model.ItemID) ([]model.ProcessID, error) {
	entry, ok := idx.ProcessIndex[item]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.CoProcesses, nil
}

func (idx *RemoteIndex) ListSymbols() (map[model.Digest]string, error) {
	return idx.TypeIndex, nil
}
