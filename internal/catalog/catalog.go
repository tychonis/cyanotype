package catalog

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tychonis/cyanotype/internal/serializer"
	"github.com/tychonis/cyanotype/model/v2"
)

type Catalog interface {
	Add(symbol model.ConcreteSymbol) error
	Get(digest string) (model.ConcreteSymbol, error)
	Find(qualifier string) (model.ConcreteSymbol, error)
}

type LocalCatalog struct {
	index map[string]string
}

func NewLocalCatalog() *LocalCatalog {
	cat := &LocalCatalog{index: make(map[string]string)}
	cat.loadIndex()
	return cat
}

func (c *LocalCatalog) loadIndex() error {
	indexPath := filepath.Join(".bpc", "index")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("open index: %w", err)
	}
	lines := bytes.Split(data, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte(":"), 2)
		if len(parts) != 2 {
			return fmt.Errorf("malformed part")
		}

		dhex := string(parts[0])
		qualifier := string(parts[1])

		c.index[dhex] = qualifier
	}
	return nil
}

func (c *LocalCatalog) appendIndexItem(key string, val string) error {
	indexPath := filepath.Join(".bpc", "index")
	f, err := os.OpenFile(indexPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open index: %w", err)
	}
	defer f.Close()
	rec := key + ":" + val + "\n"
	_, err = f.Write([]byte(rec))
	if err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return f.Sync()
}

func digestToPath(digest string) string {
	folder := digest[:2]
	return filepath.Join(".bpc", "objects", folder, digest)
}

func (c *LocalCatalog) Add(item model.ConcreteSymbol) error {
	body, err := serializer.Serialize(item)
	if err != nil {
		return err
	}
	c.index[item.GetQualifier()] = item.GetDigest()
	c.appendIndexItem(item.GetQualifier(), item.GetDigest())
	return atomicWrite(digestToPath(item.GetDigest()), body, 0o644)
}

func (c *LocalCatalog) Get(digest string) (model.ConcreteSymbol, error) {
	path := digestToPath(digest)
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ret, err := serializer.Deserialize[*model.Item](body)
	if err != nil {
		return ret, err
	}
	ret.Digest = digest
	return ret, nil
}

func (c *LocalCatalog) Find(qualifier string) (model.ConcreteSymbol, error) {
	digest, ok := c.index[qualifier]
	if !ok {
		return nil, fmt.Errorf("could not find qualifier %s", qualifier)
	}
	return c.Get(digest)
}
