package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMorphKeyTypeConfiguration(t *testing.T) {
	// Test initial state
	assert.Equal(t, MorphKeyTypeInt, GetDefaultMorphKeyType())

	// Test setting UUID
	SetDefaultMorphKeyType(MorphKeyTypeUuid)
	assert.Equal(t, MorphKeyTypeUuid, GetDefaultMorphKeyType())

	// Test setting ULID
	SetDefaultMorphKeyType(MorphKeyTypeUlid)
	assert.Equal(t, MorphKeyTypeUlid, GetDefaultMorphKeyType())

	// Test setting back to int
	SetDefaultMorphKeyType(MorphKeyTypeInt)
	assert.Equal(t, MorphKeyTypeInt, GetDefaultMorphKeyType())

	// Test convenience methods
	MorphUsingUuids()
	assert.Equal(t, MorphKeyTypeUuid, GetDefaultMorphKeyType())

	MorphUsingUlids()
	assert.Equal(t, MorphKeyTypeUlid, GetDefaultMorphKeyType())

	MorphUsingInts()
	assert.Equal(t, MorphKeyTypeInt, GetDefaultMorphKeyType())
}
