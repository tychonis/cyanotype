package catalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model"
)

var ErrNoNewRevision = errors.New("catalog is up to date")

// Adhoc hardcoded remote catalog.
func NewRemoteCatalog(endpoint string, token string, module string) *Catalog {
	client := NewHTTPClient(token)
	indexId, err := getIndexID(client, endpoint, module)
	if err != nil {
		slog.Warn("Failed to get index ID", "error", err)
		return nil
	}
	idx, _ := NewAPIIndex(endpoint+"/workspace/"+indexId, client)
	slog.Debug("Created api index.", "endpoint", endpoint, "indexId", indexId)
	cat := &Catalog{
		ModuleName: module,

		catalogType: "remote",
		storage:     NewAPIStore(endpoint, client),
		index:       idx,
	}
	err = cat.updateLatestRevision()
	if err != nil {
		if err == ErrNotFound {
			slog.Info("New catalog created")
		} else {
			slog.Warn("Failed to update latest revision", "error", err)
		}
	}
	return cat
}

func getIndexID(client *http.Client, endpoint string, moduleName string) (string, error) {
	resp, err := client.Post(endpoint+"/module/"+moduleName, "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// http.statusOK or http.statusCreated.
	if resp.StatusCode/100 != 2 {
		return "", errors.New("error response getting index ID: " + resp.Status)
	}

	var result struct {
		ID int `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(result.ID), nil
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
	Repo           string           `json:"repo"`
	LatestRevision model.RevisionID `json:"latest_revision"`
	TotalItems     int              `json:"total_items"`
	TotalCoItems   int              `json:"total_coitems"`
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

func baseName(moduleName string) string {
	parts := strings.Split(moduleName, "/")
	return parts[len(parts)-1]
}

func (cat *Catalog) UploadMetadata(client *http.Client, endpoint string) error {
	latestRev, err := cat.GetLatestRevision()
	if err != nil {
		return err
	}
	statistics, err := cat.GetStatistics()
	if err != nil {
		return err
	}
	metadata := CatalogMetadata{
		Name:           baseName(cat.ModuleName),
		Repo:           "https://" + cat.ModuleName,
		LatestRevision: latestRev.Digest,
		TotalItems:     statistics.TotalItems,
		TotalCoItems:   statistics.TotalCoItems,
	}
	content, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	resp, err := client.Post(endpoint, "application/json", bytes.NewReader(content))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response saving catalog metadata: " + resp.Status)
	}
	return nil
}

func (c *Catalog) Upload(server string, token string) error {
	client := NewHTTPClient(token)
	indexID, err := getIndexID(client, server, c.ModuleName)
	if err != nil {
		return err
	}
	endpoint := server + "/workspace/" + indexID
	err = c.UploadMetadata(client, endpoint)
	if err != nil {
		return err
	}
	return UploadIndexContent(c.index.GetContent(), endpoint, client)
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

func getArtifacts(sym model.ConcreteSymbol) []*model.Artifact {
	switch s := sym.(type) {
	case *model.Item:
		return s.Content.Artifacts
	case *model.CoItem:
		return s.Content.Artifacts
	default:
		return nil
	}
}

func (c *Catalog) Pull(other *Catalog) error {
	newRevisions, err := other.GetNewerRevisions(c.latestRevision)
	if err != nil {
		return err
	}
	slog.Debug("Pulling revisions.", "count", len(newRevisions))
	if len(newRevisions) == 0 {
		return ErrNoNewRevision
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
	total := len(allSymbols)
	fmt.Printf("Processing symbols 0/%d", total)
	count := 0
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
			err = c.Add(rev, sym)
			if err != nil {
				return err
			}
			for _, artifact := range getArtifacts(sym) {
				err = c.SaveArtifact(artifact)
				if err != nil {
					return err
				}
			}
		}
		count++
		fmt.Printf("\rProcessing symbols %d/%d", count, total)
	}
	fmt.Printf("\rProcessing Finished: %d symbols processed\n", count)
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

func (c *Catalog) AddRoot(server string, token string, qualifier string) error {
	client := NewHTTPClient(token)
	indexID, err := getIndexID(client, server, c.ModuleName)
	if err != nil {
		return err
	}
	sym, err := c.FindCurrent(qualifier)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/workspace/%s/root", server, indexID)
	req := struct {
		Digest string `json:"digest"`
	}{
		Digest: sym.GetDigest(),
	}
	content, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := client.Post(endpoint, "application/json", bytes.NewReader(content))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response adding root: " + resp.Status)
	}
	return nil
}

type APIIndex struct {
	*LocalIndex
	endpoint string
	client   *http.Client
}

func loadIndex(endpoint string, client *http.Client) (*LocalIndex, error) {
	req, err := http.NewRequest("GET", endpoint+"/index", nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return NewLocalIndex(false), nil
	}
	if resp.StatusCode/100 != 2 {
		return nil, errors.New("error response loading index: " + resp.Status)
	}

	var content IndexContent
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&content)
	if err != nil {
		return nil, err
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

func NewAPIIndex(endpoint string, client *http.Client) (*APIIndex, error) {
	localCache, err := loadIndex(endpoint, client)
	if err != nil {
		slog.Warn("Failed to load remote index, using memory index", "error", err)
		localCache = NewLocalIndex(false)
	}
	return &APIIndex{
		LocalIndex: localCache,

		endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   client,
	}, nil
}

func (a *APIIndex) SaveMetadata(digest model.Digest, metadata []byte) error {
	url := fmt.Sprintf("%s/metadata/%s", a.endpoint, digest)
	req, err := http.NewRequest("POST", url, bytes.NewReader(metadata))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response saving symbol metadata: " + resp.Status)
	}
	return nil
}

func (a *APIIndex) GetMetadata(digest model.Digest) ([]byte, error) {
	url := fmt.Sprintf("%s/metadata/%s", a.endpoint, digest)
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response loading symbol metadata: " + resp.Status)
	}
	return io.ReadAll(resp.Body)
}
