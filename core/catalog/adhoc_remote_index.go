package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/core/ranker"
	"github.com/tychonis/cyanotype/model"
)

// Adhoc hardcoded remote catalog.
func NewRemoteCatalog(endpoint string, token string, tag string) *Catalog {
	client := NewHTTPClient(token)
	cat := &Catalog{
		storage: NewAPIStore(endpoint, client),
		index:   NewRemoteIndex(endpoint+"/workspace/"+tag, client),
	}
	err := cat.updateLatestRevision()
	if err != nil {
		slog.Warn("Failed to update latest revision", "error", err)
	}
	return cat
}

// Adhoc implementation for early prototyping.
// TODO: implement incremental updates. See git negotiation algorithm
type RemoteIndex struct {
	QualifierIndex map[Qualifier]QualifierIndexEntry    `json:"qualifier_index"`
	ProcessIndex   map[model.ItemID]*ProcessIndexEntry  `json:"process_index"`
	RevisionIndex  map[model.RevisionID]*model.Revision `json:"revision_index"`

	endpoint string       `json:"-"`
	client   *http.Client `json:"-"`

	revisionOrderCache map[model.RevisionID]int `json:"-"`
	orderedRevisions   []model.RevisionID       `json:"-"`
	latestRevision     model.RevisionID         `json:"-"`
}

func RemoteIndexFromLocal(l *LocalIndex) *RemoteIndex {
	return &RemoteIndex{
		QualifierIndex: l.qualifierIndex,
		ProcessIndex:   l.processIndex,
		RevisionIndex:  l.revisionIndex,
	}
}

func NewRemoteIndex(endpoint string, client *http.Client) *RemoteIndex {
	req, err := http.NewRequest("GET", endpoint+"/index", nil)
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
	idx.endpoint = endpoint
	idx.client = client
	err = idx.buildRevisionOrderCache()
	if err != nil {
		slog.Warn("Failed to build revision order cache", "error", err)
	}
	return &idx
}

func (idx *RemoteIndex) buildRevisionOrderCache() error {
	allRevisions := make([]*model.Revision, 0, len(idx.RevisionIndex))
	for _, rev := range idx.RevisionIndex {
		allRevisions = append(allRevisions, rev)
	}
	if len(allRevisions) == 0 {
		return nil
	}
	sorted, err := ranker.StableTopoRevisions(allRevisions)
	if err != nil {
		return fmt.Errorf("rank revisions: %w", err)
	}
	if idx.revisionOrderCache == nil {
		idx.revisionOrderCache = make(map[model.RevisionID]int)
	}
	for i, rev := range sorted {
		idx.revisionOrderCache[rev] = i
	}
	idx.orderedRevisions = sorted
	idx.latestRevision = sorted[len(sorted)-1]
	return nil
}

func initRemoteIndex(endpoint string, client *http.Client) *RemoteIndex {
	ret := &RemoteIndex{
		QualifierIndex: make(map[Qualifier]QualifierIndexEntry),
		ProcessIndex:   make(map[model.ItemID]*ProcessIndexEntry),
		RevisionIndex:  make(map[model.RevisionID]*model.Revision),

		endpoint: endpoint,
		client:   client,

		revisionOrderCache: make(map[model.RevisionID]int),
		orderedRevisions:   make([]model.RevisionID, 0),
	}
	ret.buildRevisionOrderCache()
	return ret
}

func (idx *RemoteIndex) GetAllRevisions() ([]model.RevisionID, error) {
	return idx.orderedRevisions, nil
}

func (idx *RemoteIndex) GetNewerRevisions(r model.RevisionID) ([]model.RevisionID, error) {
	order, ok := idx.revisionOrderCache[r]
	if !ok {
		return nil, ErrNotFound
	}
	return idx.orderedRevisions[order:], nil
}

func (idx *RemoteIndex) Save() error {
	content, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", idx.endpoint+"/index", bytes.NewReader(content))
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

func (idx *RemoteIndex) addToQualifierIndex(revision model.RevisionID, qualifier string, symDigest model.Digest) error {
	currentEntry, ok := idx.QualifierIndex[qualifier]
	if !ok {
		currentEntry = make(QualifierIndexEntry)
		idx.QualifierIndex[qualifier] = currentEntry
	}
	currentEntry[revision] = symDigest
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

func (idx *RemoteIndex) IndexSymbol(rev *model.Revision, sym model.ConcreteSymbol) error {
	err := idx.addToQualifierIndex(rev.Digest, sym.GetQualifier(), sym.GetDigest())
	if err != nil {
		return err
	}
	return idx.indexProcess(sym)
}

func (idx *RemoteIndex) FindAll(q Qualifier) ([]model.Digest, error) {
	entry, ok := idx.QualifierIndex[q]
	if !ok {
		return nil, ErrNotFound
	}
	digests := make([]model.Digest, 0, len(entry))
	for _, digest := range entry {
		digests = append(digests, digest)
	}
	return digests, nil
}

func (idx *RemoteIndex) FindCurrent(q Qualifier) (model.Digest, error) {
	entry, ok := idx.QualifierIndex[q]
	if !ok {
		return "", ErrNotFound
	}
	allRevisions := make([]model.RevisionID, 0, len(entry))
	for rev := range entry {
		allRevisions = append(allRevisions, rev)
	}
	if len(allRevisions) == 0 {
		return "", ErrNotFound
	}
	sort.SliceStable(allRevisions, func(i, j int) bool {
		return idx.CompareRevisions(allRevisions[i], allRevisions[j]) > 0
	})
	latestRevision := entry[allRevisions[0]]
	return latestRevision, nil
}

func (idx *RemoteIndex) GetAllSymbols() ([]model.Digest, error) {
	allSymbols := make([]model.Digest, 0)
	for _, entry := range idx.QualifierIndex {
		for _, digest := range entry {
			allSymbols = append(allSymbols, digest)
		}
	}
	return allSymbols, nil
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

func (idx *RemoteIndex) GetRevision(r model.RevisionID) (*model.Revision, error) {
	revision, ok := idx.RevisionIndex[r]
	if !ok {
		return nil, ErrNotFound
	}
	return revision, nil
}

func (idx *RemoteIndex) CompareRevisions(a, b model.RevisionID) int {
	revA, err := idx.GetRevision(a)
	if err != nil {
		return 0
	}
	revB, err := idx.GetRevision(b)
	if err != nil {
		return 0
	}
	// TODO: Compare based on topological order of revisions before CreatedAt.
	if revA.CreatedAt > revB.CreatedAt {
		return -1
	} else if revA.CreatedAt < revB.CreatedAt {
		return 1
	}
	return 0
}

func (idx *RemoteIndex) IndexRevision(r *model.Revision) error {
	idx.RevisionIndex[r.Digest] = r
	return idx.buildRevisionOrderCache()
}

func (idx *RemoteIndex) GetLatestRevision() (*model.Revision, error) {
	allRevisions := make([]*model.Revision, 0, len(idx.RevisionIndex))
	for _, rev := range idx.RevisionIndex {
		allRevisions = append(allRevisions, rev)
	}
	if len(allRevisions) == 0 {
		return nil, nil
	}
	sorted, err := ranker.StableTopoRevisions(allRevisions)
	if err != nil {
		return nil, fmt.Errorf("rank revisions: %w", err)
	}
	return idx.RevisionIndex[sorted[len(sorted)-1]], nil
}

type CatalogMetadata struct {
	Name           string           `json:"name"`
	LatestRevision model.RevisionID `json:"latest_revision"`
	UniqueParts    int              `json:"unique_parts"`
}

func (idx *RemoteIndex) GetCatalogMetadata() (*CatalogMetadata, error) {
	resp, err := idx.client.Get(idx.endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response")
	}

	var metadata CatalogMetadata
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (idx *RemoteIndex) SaveCatalogMetadata() error {
	metadata := CatalogMetadata{
		Name:           "placeholder",
		LatestRevision: idx.latestRevision,
	}
	content, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	resp, err := idx.client.Post(idx.endpoint, "application/json", bytes.NewReader(content))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response")
	}
	return nil
}
