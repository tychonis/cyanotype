package process_test

import (
	"testing"

	"github.com/tychonis/cyanotype/core/process"
	"github.com/tychonis/cyanotype/model"
)

func TestAbstractImplementProcessContent(t *testing.T) {
	var _ model.ProcessContent = (*process.Abstract)(nil)
}

func TestMarshallJSONForAbstract(t *testing.T) {
	testMarshallJSON(t, &process.Abstract{})
}
