package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model"
)

// Adhoc hardcoded remote catalog.
func NewRemoteCatalog(endpoint string, token string, tag string) *Catalog {
	client := NewHTTPClient(token)
	idx, err := loadIndex(endpoint+"/workspace/"+tag, client)
	if err != nil {
		slog.Warn("Failed to load remote index", "error", err)
		idx = NewLocalIndex(false)
	}
	cat := &Catalog{
		storage: NewAPIStore(endpoint, client),
		index:   idx,
	}
	err = cat.updateLatestRevision()
	if err != nil {
		slog.Warn("Failed to update latest revision", "error", err)
	}
	return cat
}

type IndexContent struct {
	QualifierIndex map[Qualifier]QualifierIndexEntry    `json:"qualifier_index"`
	ProcessIndex   map[model.ItemID]*ProcessIndexEntry  `json:"process_index"`
	RevisionIndex  map[model.RevisionID]*model.Revision `json:"revision_index"`
}

func qualifierIndexToDigestIndex(qualifierIndex map[Qualifier]QualifierIndexEntry) map[model.Digest]DigestIndexEntry {
	digestIndex := make(map[model.Digest]DigestIndexEntry)
	for qualifier, entry := range qualifierIndex {
		for revision, digest := range entry {
			entry, ok := digestIndex[digest]
			if !ok {
				entry = make(DigestIndexEntry)
				digestIndex[digest] = entry
			}
			entry[revision] = qualifier
		}
	}
	return digestIndex
}

func loadIndex(endpoint string, client *http.Client) (*LocalIndex, error) {
	req, err := http.NewRequest("GET", endpoint+"/index", nil)
	if err != nil {
		return NewLocalIndex(false), err
	}
	resp, err := client.Do(req)
	if err != nil {
		return NewLocalIndex(false), err
	}
	defer resp.Body.Close()

	var content IndexContent
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&content)
	if err != nil {
		return NewLocalIndex(false), err
	}
	idx := &LocalIndex{
		qualifierIndex: content.QualifierIndex,
		digestIndex:    qualifierIndexToDigestIndex(content.QualifierIndex),
		processIndex:   content.ProcessIndex,
		revisionIndex:  content.RevisionIndex,

		persistent: false,
	}
	err = idx.buildRevisionOrderCache()
	return idx, err
}

func UploadIndexContent(content *IndexContent, endpoint string, client *http.Client) error {
	data, err := json.Marshal(content)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", endpoint+"/index", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("failed to save index")
	}
	return nil
}

type CatalogMetadata struct {
	Name           string           `json:"name"`
	LatestRevision model.RevisionID `json:"latest_revision"`
	UniqueParts    int              `json:"unique_parts"`
}

func GetCatalogMetadata(client *http.Client, endpoint string) (*CatalogMetadata, error) {
	resp, err := client.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response getting catalog metadata: " + resp.Status)
	}

	var metadata CatalogMetadata
	err = json.NewDecoder(resp.Body).Decode(&metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (cat *Catalog) UploadMetadata(client *http.Client, endpoint string) error {
	latestRev, err := cat.GetLatestRevision()
	if err != nil {
		return err
	}
	metadata := CatalogMetadata{
		Name:           "placeholder",
		LatestRevision: latestRev.Digest,
	}
	content, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	resp, err := client.Post(endpoint, "application/json", bytes.NewReader(content))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response saving catalog metadata: " + resp.Status)
	}
	return nil
}

func (c *Catalog) Upload(server string, token string, tag string) error {
	endpoint := server + "/workspace/" + tag
	client := NewHTTPClient(token)
	err := c.UploadMetadata(client, endpoint)
	if err != nil {
		return err
	}
	return UploadIndexContent(c.index.GetContent(), endpoint+"/index", client)
}

func (c *Catalog) GetNewerRevisions(base *model.Revision) ([]*model.Revision, error) {
	var newRevisions []model.RevisionID
	var err error
	if base == nil {
		newRevisions, err = c.index.GetAllRevisions()
	} else {
		newRevisions, err = c.index.GetNewerRevisions(base.Digest)
	}
	if err != nil {
		return nil, err
	}
	ret := make([]*model.Revision, 0, len(newRevisions))
	for _, revID := range newRevisions {
		body, err := c.storage.Load(revID)
		if err != nil {
			return nil, err
		}
		rev, err := serializer.Deserialize[*model.Revision](body)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rev)
	}
	return ret, nil
}

func (c *Catalog) Pull(other *Catalog) error {
	newRevisions, err := other.GetNewerRevisions(c.latestRevision)
	if err != nil {
		return err
	}
	slog.Debug("Pulling revisions.", "count", len(newRevisions))
	if len(newRevisions) == 0 {
		return errors.New("source catalog has no newer revisions")
	}
	for _, rev := range newRevisions {
		slog.Debug("Processing revision", "revision", rev)
		err = c.index.IndexRevision(rev)
		if err != nil {
			return err
		}
		body, err := serializer.Serialize(rev)
		if err != nil {
			return err
		}
		err = c.storage.Save(rev.Digest, body)
		if err != nil {
			return err
		}
	}
	slog.Debug("Getting symbols list.")
	allSymbols, err := other.index.GetAllSymbols()
	if err != nil {
		return err
	}
	slog.Debug("Getting symbols.")
	for _, symDigest := range allSymbols {
		sym, err := other.Get(symDigest)
		if err != nil {
			return err
		}
		metadata, err := other.GetMetadata(symDigest)
		if err != nil {
			return err
		}
		if c.latestRevision == nil || other.index.CompareRevisions(metadata.IntroducedBy, c.latestRevision.Digest) > 0 {
			revData, err := other.storage.Load(metadata.IntroducedBy)
			if err != nil {
				return err
			}
			rev, err := serializer.Deserialize[*model.Revision](revData)
			if err != nil {
				return err
			}
			c.Add(rev, sym)
		}
	}
	slog.Debug("Updating latest revision.")
	return c.updateLatestRevision()
}

func (c *Catalog) Push(other *Catalog) error {
	return other.Pull(c)
}

func (c *Catalog) updateLatestRevision() error {
	slog.Debug("Getting latest revision.")
	latestRev, err := c.index.GetLatestRevision()
	if err != nil {
		return err
	}
	if latestRev == nil {
		return nil
	}
	slog.Debug("Found latest revision", "digest", latestRev.Digest)
	fullRevData, err := c.storage.Load(latestRev.Digest)
	if err != nil {
		return err
	}
	fullRev, err := serializer.Deserialize[*model.Revision](fullRevData)
	if err != nil {
		return err
	}
	c.latestRevision = fullRev
	return nil
}
