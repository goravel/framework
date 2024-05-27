package json

import "github.com/goravel/framework/foundation"

// Marshal serializes the given value to a JSON-encoded byte slice.
func Marshal(v any) ([]byte, error) {
	return foundation.NewJSON().Marshal(v)
}

// Unmarshal deserializes the given JSON-encoded byte slice into the provided value.
func Unmarshal(data []byte, v any) error {
	return foundation.NewJSON().Unmarshal(data, v)
}

// MarshalString serializes the given value to a JSON-encoded string.
func MarshalString(v any) (string, error) {
	return foundation.NewJSON().MarshalString(v)
}

// UnmarshalString deserializes the given JSON-encoded string into the provided value.
func UnmarshalString(data string, v any) error {
	return foundation.NewJSON().UnmarshalString(data, v)
}
