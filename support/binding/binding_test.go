package binding

import (
	"testing"

	"github.com/goravel/framework/contracts/binding"
	"github.com/stretchr/testify/assert"
)

func TestDependencies(t *testing.T) {
	dependencies := Dependencies([]string{
		binding.Orm,
		binding.DB,
		binding.Schema,
		binding.Seeder,
	}...)

	assert.ElementsMatch(t, []string{
		binding.Log,
	}, dependencies)
}
