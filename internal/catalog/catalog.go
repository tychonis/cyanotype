package catalog

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/tychonis/cyanotype/internal/digest"
	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model/v2"
)

type Catalog interface {
	AddItem(item *model.Item) error
	GetItem(id model.ItemID) (*model.Item, error)
}

type LocalCatalog struct {
	index map[uuid.UUID]string
}

func (c *LocalCatalog) appendIndexItem(key uuid.UUID, val string) error {
	indexPath := filepath.Join(".bpc", "index")
	f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open index: %w", err)
	}
	defer f.Close()
	rec := key.String() + ":" + val + "\n"
	_, err = f.Write([]byte(rec))
	if err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return f.Sync()
}

func digestToPath(digest string) string {
	folder := digest[:2]
	return filepath.Join(".bpc", folder, digest)
}

func (c *LocalCatalog) AddItem(item *model.Item) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	body, err := serializer.SerializeItem(item)
	if err != nil {
		return err
	}
	item.Digest, err = digest.SHA256FromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.index[item.ID] = item.Digest
	c.appendIndexItem(item.ID, item.Digest)
	return atomicWrite(digestToPath(item.Digest), body, 0o644)
}

func (c *LocalCatalog) GetItem(id model.ItemID) (*model.Item, error) {
	digest, ok := c.index[id]
	if !ok {
		return nil, errors.New("not found")
	}
	path := digestToPath(digest)
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ret, err := serializer.DeserializeItem(body)
	if err != nil {
		return ret, err
	}
	ret.Digest = digest
	return ret, nil
}
