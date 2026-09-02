package event

import (
	stderrors "errors"
	"slices"

	"github.com/goravel/framework/contracts/event"
)

var _ event.Result = (*Result)(nil)

// Result collects the errors returned by the listeners of a single dispatch.
type Result struct {
	errs []error
}

// newResult creates the result of a single event dispatch.
func newResult(errs []error) *Result {
	return &Result{errs: errs}
}

func (r *Result) Error() error {
	return stderrors.Join(r.errs...)
}

func (r *Result) Errors() []error {
	return slices.Clone(r.errs)
}

func (r *Result) Failed() bool {
	return len(r.errs) > 0
}
