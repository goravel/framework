package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModuleName(t *testing.T) {
	assert.Equal(t, "github.com/goravel/framework", ModuleName())
}

func TestModuleNameFromArgs(t *testing.T) {
	assert.Equal(t, "test", ModuleNameFromArgs([]string{"go", "run", ".", "--module=test"}))
	assert.Equal(t, "goravel", ModuleNameFromArgs([]string{"go", "run", "."}))
}
