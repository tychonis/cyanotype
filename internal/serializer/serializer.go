package serializer

import (
	"encoding/json"
)

func Serialize(s any) ([]byte, error) {
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

func Deserialize[T any](body []byte) (T, error) {
	var ret T
	err := json.Unmarshal(body, &ret)
	return ret, err
}
