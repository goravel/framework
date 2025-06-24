package debug

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncInfo(t *testing.T) {
	var f = func() {}
	info := GetFuncInfo(f)
	assert.Equal(t, "github.com/goravel/framework/support/debug.TestFuncInfo.func1", info.Name)
	assert.Equal(t, "TestFuncInfo.func1", info.ShortName())
	assert.Equal(t, "github.com/goravel/framework/support/debug", info.PackagePath())
	assert.Equal(t, "debug", info.PackageName())
	assert.Contains(t, info.File, "func_test.go")
	assert.Equal(t, 10, info.Line)

	info = GetFuncInfo(Dump)
	assert.Equal(t, "github.com/goravel/framework/support/debug.Dump", info.Name)
	assert.Equal(t, "Dump", info.ShortName())
	assert.Equal(t, "github.com/goravel/framework/support/debug", info.PackagePath())
	assert.Equal(t, "debug", info.PackageName())
	assert.Contains(t, info.File, "dump.go")
	assert.Equal(t, 11, info.Line)
}
