package process_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tychonis/cyanotype/model"
)

func testMarshallJSON[T model.ProcessContent](t *testing.T, sample T) {
	data, err := json.Marshal(&sample)
	if err != nil {
		t.Error("failed to marshal content.", "type", sample.GetType(), "error", err)
	}
	fmt.Print(string(data))
	var m map[string]string
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Error("failed to unmarshal content.", "type", sample.GetType(), "error", err)
	}
	if m["type"] != sample.GetType() {
		t.Error("wrong type.", "expected", sample.GetType(), "actual", m["type"])
	}
}
