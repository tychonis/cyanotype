package serializer

import (
	"encoding/json"

	"github.com/tychonis/cyanotype/model/v2"
)

func SerializeItem(item *model.Item) ([]byte, error) {
	return json.Marshal(item)
}
