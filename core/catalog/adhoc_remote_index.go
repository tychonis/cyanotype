package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/model"
)

// Adhoc implementation for early prototyping.
// TODO: implement incremental updates. See git negotiation algorithm
type RemoteIndex struct {
	QualifierIndex map[Qualifier]model.ItemID          `json:"qualifier_index"`
	ProcessIndex   map[model.ItemID]*ProcessIndexEntry `json:"process_index"`

	Endpoint string `json:"-"`

	client *http.Client `json:"-"`
}

func RemoteIndexFromLocal(l *LocalIndex) *RemoteIndex {
	return &RemoteIndex{
		QualifierIndex: l.qualifierIndex,
		ProcessIndex:   l.processIndex,
	}
}

func NewRemoteIndex(endpoint string, client *http.Client) *RemoteIndex {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return initRemoteIndex(endpoint, client)
	}
	resp, err := client.Do(req)
	if err != nil {
		return initRemoteIndex(endpoint, client)
	}
	defer resp.Body.Close()

	var idx RemoteIndex
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&idx)
	if err != nil {
		return initRemoteIndex(endpoint, client)
	}
	return &idx
}

func initRemoteIndex(endpoint string, client *http.Client) *RemoteIndex {
	return &RemoteIndex{
		QualifierIndex: make(map[Qualifier]model.ItemID),
		ProcessIndex:   make(map[model.ItemID]*ProcessIndexEntry),

		Endpoint: endpoint,

		client: client,
	}
}

func (idx *RemoteIndex) Save() error {
	content, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", idx.Endpoint, bytes.NewReader(content))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := idx.client.Do(req)
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
	case *process.Process:
		for _, bomLine := range resolved.Input() {
			idx.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output() {
			idx.addToProcessIndex("process", bomLine.Item, resolved.Digest)
		}
	case *process.CoProcess:
		for _, bomLine := range resolved.Input() {
			idx.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
		for _, bomLine := range resolved.Output() {
			idx.addToProcessIndex("coprocess", bomLine.Item, resolved.Digest)
		}
	}
	return nil
}

func (idx *RemoteIndex) Index(sym model.ConcreteSymbol) error {
	err := idx.addToMainIndex(sym.GetQualifier(), sym.GetDigest())
	if err != nil {
		return err
	}
	return idx.indexProcess(sym)
}

func (idx *RemoteIndex) FindCurrent(q Qualifier) (model.Digest, error) {
	digest, ok := idx.QualifierIndex[q]
	if !ok {
		return "", ErrNotFound
	}
	return digest, nil
}

func (idx *RemoteIndex) GetItemProcesses(item model.ItemID) ([]process.ProcessID, error) {
	entry, ok := idx.ProcessIndex[item]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.Processes, nil
}

func (idx *RemoteIndex) GetItemCoProcesses(item model.ItemID) ([]process.ProcessID, error) {
	entry, ok := idx.ProcessIndex[item]
	if !ok {
		return nil, ErrNotFound
	}
	return entry.CoProcesses, nil
}
