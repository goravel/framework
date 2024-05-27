package support

type Json interface {
	// Marshal returns the JSON encoding of v.
	Marshal(any) ([]byte, error)
	// MarshalString returns the JSON encoding of v as a string.
	MarshalString(any) (string, error)
	// Unmarshal parses the JSON-encoded data and stores the result in the value pointed to by v.
	Unmarshal([]byte, any) error
	// UnmarshalString parses the JSON-encoded data and stores the result in the value pointed to by v.
	UnmarshalString(string, any) error
}
