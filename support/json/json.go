package json

import "github.com/goravel/framework/foundation"

// Marshal serializes the given value to a JSON-encoded byte slice.
func Marshal(v any) ([]byte, error) {
	return foundation.App.GetJSON().Marshal(v)
}

// Unmarshal deserializes the given JSON-encoded byte slice into the provided value.
func Unmarshal(data []byte, v any) error {
	return foundation.App.GetJSON().Unmarshal(data, v)
}
