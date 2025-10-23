package catalog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tychonis/cyanotype/model"
)

type Storage interface {
	Save(digest model.Digest, data []byte) error
	Load(digest model.Digest) ([]byte, error)
}

type LocalFile struct{}

func digestToPath(digest string) string {
	folder := digest[:2]
	return filepath.Join(".bpc", "objects", folder, digest)
}

func (f *LocalFile) Save(digest model.Digest, data []byte) error {
	path := digestToPath(digest)
	return atomicWrite(path, data, 0o644)
}

func (f *LocalFile) Load(digest model.Digest) ([]byte, error) {
	path := digestToPath(digest)
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

type APIStore struct {
	endpoint string
}

func NewAPIStore(endpoint string) *APIStore {
	return &APIStore{
		endpoint: endpoint,
	}
}

func (a *APIStore) Save(digest model.Digest, data []byte) error {
	url := fmt.Sprintf("%s/%s", a.endpoint, digest)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return errors.New("error response")
	}
	return nil
}

func (a *APIStore) Load(digest model.Digest) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", a.endpoint, digest)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response")
	}
	return io.ReadAll(resp.Body)
}
