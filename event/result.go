package event

import (
	stderrors "errors"

	"github.com/goravel/framework/contracts/event"
)

var _ event.Result = (*Result)(nil)

// Result collects the errors returned by the listeners of a single dispatch.
type Result struct {
	errs []error
}

func NewResult(errs []error) *Result {
	return &Result{errs: errs}
}

func (r *Result) Error() error {
	return stderrors.Join(r.errs...)
}

func (r *Result) Errors() []error {
	return r.errs
}

func (r *Result) Failed() bool {
	return len(r.errs) > 0
}
