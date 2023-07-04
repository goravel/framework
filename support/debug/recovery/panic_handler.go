package recovery

import (
	"github.com/goravel/framework/contracts/support/debug/recovery"
)

type PanicHandler struct{}

var _ recovery.Handler = (*PanicHandler)(nil)

func (h *PanicHandler) ShouldReport(v interface{}) bool {
	return true
}

func (h *PanicHandler) Report(v interface{}) {
	panic(v)
}
