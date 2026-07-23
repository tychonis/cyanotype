package instantiator

import (
	"errors"

	"github.com/tychonis/cyanotype/core/catalog"
	"github.com/tychonis/cyanotype/model"
)

type ItemVersion struct {
	Item     *model.Item     `json:"item"`
	Revision *model.Revision `json:"revision"`
}

// TODO: Consider moving this to a more appropriate location.
func History(cat *catalog.Catalog, coItem model.ItemID) ([]*ItemVersion, error) {
	cps, err := cat.GetItemCoProcesses(coItem)
	if err != nil {
		return nil, err
	}
	ret := make([]*ItemVersion, 0, len(cps))
	for _, cp := range cps {
		input := cp.Input()
		if len(input) == 1 {
			itemID := input[0].Item
			itemSym, err := cat.Get(itemID)
			if err != nil {
				continue
			}
			item, ok := itemSym.(*model.Item)
			if !ok {
				return nil, errors.New("invalid item type")
			}
			itemMetadata, err := cat.GetMetadata(itemID)
			if err != nil {
				continue
			}
			lastCommitted := itemMetadata.LastCommitted()
			revision, err := cat.GetRevision(lastCommitted)
			if err != nil {
				continue
			}
			ret = append(ret, &ItemVersion{
				Item:     item,
				Revision: revision,
			})
		}

	}
	return ret, nil
}
