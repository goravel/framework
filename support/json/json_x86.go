//go:build amd64

package json

import (
	"encoding/json"

	"github.com/bytedance/sonic"

	"github.com/goravel/framework/contracts/support"
)

type Json struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

func New() support.Json {
	return &Json{
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}
}

func (j *Json) Marshal(v any) ([]byte, error) {
	return j.marshal(v)
}

func (j *Json) MarshalString(v any) (string, error) {
	s, err := j.marshal(v)
	return string(s), err
}

func (j *Json) Unmarshal(data []byte, v any) error {
	return j.unmarshal(data, v)
}

func (j *Json) UnmarshalString(data string, v any) error {
	return j.unmarshal([]byte(data), v)
}

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
