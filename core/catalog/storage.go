package catalog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tychonis/cyanotype/model"
)

type Storage interface {
	Save(digest model.Digest, data []byte) error
	SaveMetadata(digest model.Digest, metadata []byte) error
	Load(digest model.Digest) ([]byte, error)
	LoadMetadata(digest model.Digest) ([]byte, error)
}

type LocalStorage struct{}

func digestToPath(digest string) (string, error) {
	if len(digest) < 2 {
		return "", errors.New("incorrect digest")
	}
	folder := digest[:2]
	return filepath.Join(".bpc", "objects", folder, digest), nil
}

func (ls *LocalStorage) Save(digest model.Digest, data []byte) error {
	path, err := digestToPath(digest)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return atomicWrite(path, data, 0o644)
}

func (ls *LocalStorage) SaveMetadata(digest model.Digest, metadata []byte) error {
	path, err := digestToPath(digest)
	if err != nil {
		return err
	}

	path = path + ".meta"
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return atomicWrite(path, metadata, 0o644)
}

func (ls *LocalStorage) Load(digest model.Digest) ([]byte, error) {
	path, err := digestToPath(digest)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

func (ls *LocalStorage) LoadMetadata(digest model.Digest) ([]byte, error) {
	path, err := digestToPath(digest)
	if err != nil {
		return nil, err
	}
	path = path + ".meta"
	return os.ReadFile(path)
}

type MemoryStore struct {
	storage map[string][]byte
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		storage: make(map[string][]byte),
	}
}

func (m *MemoryStore) Save(digest model.Digest, data []byte) error {
	m.storage[digest] = data
	return nil
}

func (m *MemoryStore) Load(digest model.Digest) ([]byte, error) {
	data, ok := m.storage[digest]
	if !ok {
		return nil, errors.New("digest not found")
	}
	return data, nil
}

func (m *MemoryStore) SaveMetadata(digest model.Digest, metadata []byte) error {
	m.storage[digest+".meta"] = metadata
	return nil
}

func (m *MemoryStore) LoadMetadata(digest model.Digest) ([]byte, error) {
	metadata, ok := m.storage[digest+".meta"]
	if !ok {
		return nil, errors.New("metadata not found")
	}
	return metadata, nil
}

type APIStore struct {
	endpoint string

	client *http.Client
}

func NewAPIStore(endpoint string, client *http.Client) *APIStore {
	return &APIStore{
		endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   client,
	}
}

func (a *APIStore) Save(digest model.Digest, data []byte) error {
	url := fmt.Sprintf("%s/definition/%s", a.endpoint, digest)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response")
	}
	return nil
}

func (a *APIStore) Load(digest model.Digest) ([]byte, error) {
	url := fmt.Sprintf("%s/definition/%s", a.endpoint, digest)
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response")
	}
	return io.ReadAll(resp.Body)
}

func (a *APIStore) SaveMetadata(digest model.Digest, metadata []byte) error {
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
		return errors.New("error response")
	}
	return nil
}

func (a *APIStore) LoadMetadata(digest model.Digest) ([]byte, error) {
	url := fmt.Sprintf("%s/metadata/%s", a.endpoint, digest)
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response")
	}
	return io.ReadAll(resp.Body)
}
