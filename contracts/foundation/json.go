package foundation

type Json interface {
	// Marshal serializes the given value to a JSON-encoded byte slice.
	Marshal(any) ([]byte, error)
	// Unmarshal deserializes the given JSON-encoded byte slice into the provided value.
	Unmarshal([]byte, any) error
}
