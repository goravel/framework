package json

import (
	"encoding/json"

	"github.com/bytedance/sonic"
	"github.com/goravel/framework/support/env"
)

func Marshal(v any) ([]byte, error) {
	if env.IsX86() {
		return sonic.Marshal(v)
	} else {
		return json.Marshal(v)
	}
}

func Unmarshal(data []byte, v any) error {
	if env.IsX86() {
		return sonic.Unmarshal(data, v)
	} else {
		return json.Unmarshal(data, v)
	}
}

func MarshalString(v any) (string, error) {
	if env.IsX86() {
		return sonic.MarshalString(v)
	} else {
		s, err := json.Marshal(v)
		return string(s), err
	}
}

func UnmarshalString(data string, v any) error {
	if env.IsX86() {
		return sonic.UnmarshalString(data, v)
	} else {
		return json.Unmarshal([]byte(data), v)
	}
}
