package stable

import (
	"bytes"
	"encoding/json"
	"sort"
)

// Map is a JSON object with deterministic serialization.
type Map map[string]any

// MarshalJSON implements json.Marshaler.
// It sorts keys to ensure deterministic output.
func (m Map) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := &bytes.Buffer{}
	buf.WriteByte('{')

	for i, k := range keys {
		// Encode key
		keyBytes, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(keyBytes)
		buf.WriteByte(':')

		// Encode value
		valBytes, err := json.Marshal(m[k])
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)

		// Add comma if not last
		if i < len(keys)-1 {
			buf.WriteByte(',')
		}
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}
