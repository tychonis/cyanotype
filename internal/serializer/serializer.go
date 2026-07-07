package serializer

import (
	"encoding/json"

	"github.com/tychonis/cyanotype/model"
)

func Serialize(s model.Symbol) ([]byte, error) {
	return json.Marshal(s)
}

func GetType(body []byte) (string, error) {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return "", err
	}
	return probe.Type, nil
}

func Deserialize[T model.Symbol](body []byte) (T, error) {
	var ret T
	err := json.Unmarshal(body, &ret)
	return ret, err
}
