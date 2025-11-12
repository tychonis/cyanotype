package cerror

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

type ErrWithRange struct {
	r   hcl.Range
	msg string
}

func (e *ErrWithRange) Error() string {
	return fmt.Sprintf("%s:%d %s", e.r.Filename, e.r.Start.Line, e.msg)
}

func ErrorWithRange(msg string, r hcl.Range) *ErrWithRange {
	return &ErrWithRange{
		r:   r,
		msg: msg,
	}
}
