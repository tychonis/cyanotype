package stable_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/tychonis/cyanotype/internal/stable"
)

func TestStableMap_NilIsNull(t *testing.T) {
	var m stable.Map = nil
	got, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	if string(got) != "null" {
		t.Fatalf("want null, got %s", string(got))
	}
}

func TestStableMap_EmptyIsEmptyObject(t *testing.T) {
	m := stable.Map{}
	got, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	if string(got) != "{}" {
		t.Fatalf("want {}, got %s", string(got))
	}
}

func TestStableMap_SortedKeys(t *testing.T) {
	m := stable.Map{
		"z": 1,
		"a": 2,
		"b": "x",
	}
	got, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}
	want := `{"a":2,"b":"x","z":1}`
	if string(got) != want {
		t.Fatalf("want %s, got %s", want, string(got))
	}
}

func TestStableMap_RepeatableAcrossRuns(t *testing.T) {
	// Create a map with shuffled insertion order
	const n = 50
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = fmt.Sprintf("k%03d", i)
	}
	// Deterministic shuffle
	r := rand.New(rand.NewSource(42))
	r.Shuffle(n, func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	m := stable.Map{}
	for i, k := range keys {
		m[k] = i
	}

	first, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Marshal multiple times; result must be byte-for-byte identical.
	for i := 0; i < 20; i++ {
		got, err := m.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON error (iter %d): %v", i, err)
		}
		if string(got) != string(first) {
			t.Fatalf("non-deterministic output at iter %d:\nfirst=%s\ngot  =%s",
				i, string(first), string(got))
		}
	}

	// Also verify it equals the expected sorted form.
	sortedKeys := append([]string(nil), keys...)
	sort.Strings(sortedKeys)
	// Build expected JSON via stdlib on a sorted slice to avoid manual formatting mistakes.
	type obj map[string]any
	exp := obj{}
	for _, k := range sortedKeys {
		// value is the original index of k in the shuffled list
		// Find i in the shuffled list
		var val int
		for idx, kk := range keys {
			if kk == k {
				val = idx
				break
			}
		}
		exp[k] = val
	}
	want, _ := json.Marshal(exp)
	if string(first) != string(want) {
		t.Fatalf("want %s, got %s", string(want), string(first))
	}
}

func TestStableMap_NestedMapIsValidJSON(t *testing.T) {
	// NOTE: Nested plain map[string]any is NOT guaranteed to be ordered by StableMap.
	// We only assert it’s valid JSON and round-trips to equivalent structure.
	m := stable.Map{
		"a": 1,
		"b": map[string]any{"y": 2, "x": 3},
	}
	raw, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal produced JSON failed: %v; json=%s", err, string(raw))
	}

	// Compare top-level structure.
	if _, ok := got["a"]; !ok {
		t.Fatalf("missing key a in %v", got)
	}
	// Nested map should contain both x and y (order not checked).
	b, ok := got["b"].(map[string]any)
	if !ok {
		// Depending on decoder, it may be map[string]interface{} already; reflect handles both.
		if rb := reflect.ValueOf(got["b"]); rb.Kind() == reflect.Map {
			// good enough
		} else {
			t.Fatalf("nested b is not an object: %T", got["b"])
		}
	} else {
		if _, ok := b["x"]; !ok {
			t.Fatalf("nested map missing x")
		}
		if _, ok := b["y"]; !ok {
			t.Fatalf("nested map missing y")
		}
	}
}

func TestStableMap_SpecialCharsInKeysAndValues(t *testing.T) {
	m := stable.Map{
		`quote"key`: `val"ue`,
		"newline\n": "line\nbreak",
		"tab\t":     "\t",
		"unicode":   "漢字",
	}

	got, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	// Ensure it’s valid JSON and round-trips losslessly.
	var round map[string]any
	if err := json.Unmarshal(got, &round); err != nil {
		t.Fatalf("invalid JSON produced: %v; json=%s", err, string(got))
	}

	// The stdlib unescapes strings on Unmarshal, so we can compare values directly.
	for k, v := range m {
		if round[k] != v {
			t.Fatalf("round-trip mismatch for key %q: want %q, got %q", k, v, round[k])
		}
	}
}

// Optional: a micro-benchmark to confirm perf isn’t terrible.
func BenchmarkStableMapMarshalJSON(b *testing.B) {
	m := stable.Map{}
	for i := 0; i < 500; i++ {
		m[fmt.Sprintf("k%04d", i)] = i
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = m.MarshalJSON()
	}
}

// Ensure time import isn’t unused if you decide to seed randomness differently later.
var _ = time.Now
