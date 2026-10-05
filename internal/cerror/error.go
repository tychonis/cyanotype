package cerror

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type ErrWithRange struct {
	Range hcl.Range
	Err   error
}

func (e *ErrWithRange) Error() string {
	return fmt.Sprintf("%s:%d %s", e.Range.Filename, e.Range.Start.Line, e.Err)
}

func ErrorWithRange(err error, r hcl.Range) *ErrWithRange {
	return &ErrWithRange{
		Range: r,
		Err:   err,
	}
}

type ErrKeyNotFound struct {
	Key string
}

func (e *ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key not found: %s", e.Key)
}

func KeyNotFoundError(key string) *ErrKeyNotFound {
	return &ErrKeyNotFound{
		Key: key,
	}
}
