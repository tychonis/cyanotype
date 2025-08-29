package catalog

import "github.com/tychonis/cyanotype/model"

type Catalog interface {
	AddItem(item *model.Item) error
	FindItem(qualifier string) (model.ItemID, bool)
	GetItem(id model.ItemID) (*model.Item, error)
}
