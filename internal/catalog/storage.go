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

func (ls *LocalStorage) Load(digest model.Digest) ([]byte, error) {
	path, err := digestToPath(digest)
	if err != nil {
		return nil, err
	}
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
	token    string
}

func NewAPIStore(endpoint string, token string) *APIStore {
	return &APIStore{
		endpoint: endpoint,
		token:    token,
	}
}

func (a *APIStore) Save(digest model.Digest, data []byte) error {
	url := fmt.Sprintf("%s/%s", a.endpoint, digest)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if a.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
	}

	resp, err := http.DefaultClient.Do(req)
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
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if a.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("error response")
	}
	return io.ReadAll(resp.Body)
}
