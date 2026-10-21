package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFake(t *testing.T) {
	fake := NewFake()
	assert.NotNil(t, fake)
	assert.NotNil(t, fake.Instance())
	assert.NotEmpty(t, fake.Name())
	assert.NotEmpty(t, fake.Email())
	assert.NoError(t, fake.Seed(42))
}
