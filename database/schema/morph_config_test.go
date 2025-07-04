package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMorphKeyTypeConfiguration(t *testing.T) {
	// Test initial state
	assert.Equal(t, MorphKeyTypeInt, GetDefaultMorphKeyType())

	// Test setting UUID
	err := SetDefaultMorphKeyType(MorphKeyTypeUuid)
	assert.NoError(t, err)
	assert.Equal(t, MorphKeyTypeUuid, GetDefaultMorphKeyType())

	// Test setting ULID
	err = SetDefaultMorphKeyType(MorphKeyTypeUlid)
	assert.NoError(t, err)
	assert.Equal(t, MorphKeyTypeUlid, GetDefaultMorphKeyType())

	// Test setting back to int
	err = SetDefaultMorphKeyType(MorphKeyTypeInt)
	assert.NoError(t, err)
	assert.Equal(t, MorphKeyTypeInt, GetDefaultMorphKeyType())

	// Test invalid key type
	err = SetDefaultMorphKeyType("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")

	// Test convenience methods
	MorphUsingUuids()
	assert.Equal(t, MorphKeyTypeUuid, GetDefaultMorphKeyType())

	MorphUsingUlids()
	assert.Equal(t, MorphKeyTypeUlid, GetDefaultMorphKeyType())

	MorphUsingInts()
	assert.Equal(t, MorphKeyTypeInt, GetDefaultMorphKeyType())
}