package process_test

import (
	"testing"

	"github.com/tychonis/cyanotype/core/process"
)

func TestAbstractImplementProcessContent(t *testing.T) {
	var _ process.ProcessContent = (*process.Abstract)(nil)
}

func TestMarshallJSONForAbstract(t *testing.T) {
	testMarshallJSON(t, &process.Abstract{})
}
