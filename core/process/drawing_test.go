package process_test

import (
	"testing"

	"github.com/tychonis/cyanotype/core/process"
)

func TestDrawingImplementProcessContent(t *testing.T) {
	var _ process.ProcessContent = (*process.Drawing)(nil)
}

func TestMarshallJSONForDrawing(t *testing.T) {
	testMarshallJSON(t, &process.Drawing{})
}
