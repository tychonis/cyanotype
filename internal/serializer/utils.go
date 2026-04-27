package serializer

import "encoding/json"

func JSONWithKey(obj any, key string, val string) ([]byte, error) {
	m := map[string]any{
		key: val,
	}
	b, _ := json.Marshal(obj)
	json.Unmarshal(b, &m)
	return json.Marshal(m)
}
