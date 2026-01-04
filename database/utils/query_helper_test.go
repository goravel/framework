package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
)

func TestPrepareWhereOperatorAndValue(t *testing.T) {
	t.Run("with two arguments - operator and value", func(t *testing.T) {
		op, value, err := PrepareWhereOperatorAndValue(">", 100)
		assert.Nil(t, err)
		assert.Equal(t, ">", op)
		assert.Equal(t, 100, value)
	})

	t.Run("with one argument - defaults to equals operator", func(t *testing.T) {
		op, value, err := PrepareWhereOperatorAndValue("John")
		assert.Nil(t, err)
		assert.Equal(t, "=", op)
		assert.Equal(t, "John", value)
	})

	t.Run("with no arguments - returns error", func(t *testing.T) {
		op, value, err := PrepareWhereOperatorAndValue()
		assert.NotNil(t, err)
		assert.Nil(t, op)
		assert.Nil(t, value)
		assert.Equal(t, errors.DatabaseInvalidArgumentNumber.Args(0, "1 or 2"), err)
	})

	t.Run("with more than two arguments - returns error", func(t *testing.T) {
		op, value, err := PrepareWhereOperatorAndValue(">", 100, "extra")
		assert.NotNil(t, err)
		assert.Nil(t, op)
		assert.Nil(t, value)
		assert.Equal(t, errors.DatabaseInvalidArgumentNumber.Args(3, "1 or 2"), err)
	})
}
