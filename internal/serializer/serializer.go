package serializer

import (
	"encoding/json"

	"github.com/tychonis/cyanotype/model"
)

func Serialize(s model.Typed) ([]byte, error) {
	type wrapper struct {
		Type string `json:"type"`
		Data any    `json:"data"`
	}

	return json.Marshal(wrapper{
		Type: s.GetType(),
		Data: s,
	})
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

func Deserialize[T any](body []byte) (T, error) {
	var wrapper struct {
		Data T `json:"data"`
	}
	err := json.Unmarshal(body, &wrapper)
	return wrapper.Data, err
}
