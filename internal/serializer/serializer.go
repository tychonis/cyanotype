package serializer

import (
	"encoding/json"

	"github.com/tychonis/cyanotype/model/v2"
)

func SerializeItem(item *model.Item) ([]byte, error) {
	return json.Marshal(item)
}

func DeserializeItem(body []byte) (*model.Item, error) {
	ret := model.Item{}
	err := json.Unmarshal(body, &ret)
	return &ret, err
}

func Serialize(s model.Symbol) ([]byte, error) {
	return json.Marshal(s)
}

func Deserialize[T model.Symbol](body []byte) (T, error) {
	var ret T
	err := json.Unmarshal(body, &ret)
	return ret, err
}
