package gorm

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRow_Err returns the stored error from the Row.
func TestRow_Err_Nil(t *testing.T) {
	r := &Row{err: nil}
	assert.NoError(t, r.Err())
}

func TestRow_Err_NonNil(t *testing.T) {
	expected := errors.New("scan error")
	r := &Row{err: expected}
	assert.Same(t, expected, r.Err())
}
