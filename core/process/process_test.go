package process_test

import (
	"encoding/json"
	"testing"

	"github.com/tychonis/cyanotype/core/process"
)

func testMarshallJSON[T process.ProcessContent](t *testing.T, sample T) {
	data, err := json.Marshal(&sample)
	if err != nil {
		t.Error("failed to marshal content.", "type", sample.GetType(), "error", err)
	}
	var m map[string]string
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Error("failed to unmarshal content.", "type", sample.GetType(), "error", err)
	}
	if err != nil || m["type"] != sample.GetType() {
		t.Error("wrong type.", "expected", sample.GetType(), "actual", m["type"])
	}

	p := &process.Process{}
	p.Content = sample

	data, err = json.Marshal(p)
	if err != nil {
		t.Error("failed to marshal process.", "content_type", sample.GetType(), "error", err)
	}
	var dp process.Process
	err = json.Unmarshal(data, &dp)
	if err != nil || dp.Content == nil {
		t.Error("failed to unmarshal process.",
			"content_type", sample.GetType(), "error", err,
			"expected", sample.GetType(), "actual", nil,
		)
		return
	}
	if dp.Content.GetType() != sample.GetType() {
		t.Error("failed to unmarshal process.",
			"content_type", sample.GetType(), "error", err,
			"expected", sample.GetType(), "actual", dp.Content.GetType(),
		)
	}
}
