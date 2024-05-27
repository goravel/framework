//go:build amd64

package json

import (
	"github.com/bytedance/sonic"
)

// Marshal is a wrapper of sonic.Marshal.
// Marshal 是 sonic.Marshal 的包装器。
func Marshal(v any) ([]byte, error) {
	return sonic.Marshal(v)
}

// Unmarshal is a wrapper of sonic.Unmarshal.
// Unmarshal 是 sonic.Unmarshal 的包装器。
func Unmarshal(data []byte, v any) error {
	return sonic.Unmarshal(data, v)
}

// MarshalString is a wrapper of sonic.MarshalString.
// MarshalString 是 sonic.MarshalString 的包装器。
func MarshalString(v any) (string, error) {
	return sonic.MarshalString(v)
}

// UnmarshalString is a wrapper of sonic.UnmarshalString.
// UnmarshalString 是 sonic.UnmarshalString 的包装器。
func UnmarshalString(data string, v any) error {
	return sonic.UnmarshalString(data, v)
}
