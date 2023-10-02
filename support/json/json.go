package json

import (
	"encoding/json"

	"github.com/bytedance/sonic"
	"github.com/goravel/framework/support/env"
)

// Marshal is a wrapper of json.Marshal or sonic.Marshal.
// Marshal 是 json.Marshal 或 sonic.Marshal 的包装器。
func Marshal(v any) ([]byte, error) {
	if env.IsX86() {
		return sonic.Marshal(v)
	} else {
		return json.Marshal(v)
	}
}

// Unmarshal is a wrapper of json.Unmarshal or sonic.Unmarshal.
// Unmarshal 是 json.Unmarshal 或 sonic.Unmarshal 的包装器。
func Unmarshal(data []byte, v any) error {
	if env.IsX86() {
		return sonic.Unmarshal(data, v)
	} else {
		return json.Unmarshal(data, v)
	}
}

// MarshalString is a wrapper of json.Marshal or sonic.MarshalString.
// MarshalString 是 json.Marshal 或 sonic.MarshalString 的包装器。
func MarshalString(v any) (string, error) {
	if env.IsX86() {
		return sonic.MarshalString(v)
	} else {
		s, err := json.Marshal(v)
		return string(s), err
	}
}

// UnmarshalString is a wrapper of json.Unmarshal or sonic.UnmarshalString.
// UnmarshalString 是 json.Unmarshal 或 sonic.UnmarshalString 的包装器。
func UnmarshalString(data string, v any) error {
	if env.IsX86() {
		return sonic.UnmarshalString(data, v)
	} else {
		return json.Unmarshal([]byte(data), v)
	}
}
