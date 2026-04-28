package process_test

import (
	"testing"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/model"
)

func TestDrawingImplementProcessContent(t *testing.T) {
	var _ model.ProcessContent = (*process.Drawing)(nil)
}

func TestMarshallJSONForDrawing(t *testing.T) {
	testMarshallJSON(t, &process.Drawing{})
}
