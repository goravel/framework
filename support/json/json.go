//go:build !amd64

package json

import (
	"encoding/json"
)

// Marshal is a wrapper of json.Marshal.
// Marshal 是 json.Marshal 的包装器。
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal is a wrapper of json.Unmarshal.
// Unmarshal 是 json.Unmarshal 的包装器。
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// MarshalString is a wrapper of json.Marshal.
// MarshalString 是 json.Marshal 的包装器。
func MarshalString(v any) (string, error) {
	s, err := json.Marshal(v)
	return string(s), err
}

// UnmarshalString is a wrapper of json.Unmarshal.
// UnmarshalString 是 json.Unmarshal 的包装器。
func UnmarshalString(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}
