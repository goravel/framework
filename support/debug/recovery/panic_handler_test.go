package recovery

import (
	"errors"
	"strings"
	"testing"

	"github.com/goravel/framework/contracts/support/debug/recovery"
	"github.com/stretchr/testify/assert"
)

type testPanicHandler struct {
	*PanicHandler
}

func (h *testPanicHandler) ShouldReport(v interface{}) bool {
	if err, ok := v.(error); ok {
		return !strings.Contains(err.Error(), "no-report")
	}

	return true
}

type kernel struct {
	recovery recovery.Handler
}

func newKernel(recovery recovery.Handler) *kernel {
	return &kernel{
		recovery: recovery,
	}
}

func (k *kernel) Handle(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			if k.recovery.ShouldReport(r) {
				k.recovery.Report(r)
			}
		}
	}()

	fn()
}

func TestPanicHandler(t *testing.T) {
	k := newKernel(&testPanicHandler{})

	assert.NotPanics(t, func() {
		k.Handle(func() {
			panic(errors.New("test-no-report"))
		})
	})

	assert.Panics(t, func() {
		k.Handle(func() {
			panic("test-report")
		})
	})
}
