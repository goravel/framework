package schema

import "fmt"

// MorphKeyType represents the type of key used for morph relationships
type MorphKeyType string

const (
	MorphKeyTypeInt  MorphKeyType = "int"
	MorphKeyTypeUuid MorphKeyType = "uuid"
)

var defaultMorphKeyType MorphKeyType = MorphKeyTypeInt

// SetDefaultMorphKeyType sets the default morph key type
func SetDefaultMorphKeyType(keyType MorphKeyType) error {
	if keyType != MorphKeyTypeInt && keyType != MorphKeyTypeUuid {
		return fmt.Errorf("morph key type must be '%s' or '%s'", MorphKeyTypeInt, MorphKeyTypeUuid)
	}
	defaultMorphKeyType = keyType
	return nil
}

// GetDefaultMorphKeyType returns the current default morph key type
func GetDefaultMorphKeyType() MorphKeyType {
	return defaultMorphKeyType
}

// MorphUsingUuids sets the default morph key type to UUID
func MorphUsingUuids() {
	defaultMorphKeyType = MorphKeyTypeUuid
}

// MorphUsingInts sets the default morph key type to int (default)
func MorphUsingInts() {
	defaultMorphKeyType = MorphKeyTypeInt
}