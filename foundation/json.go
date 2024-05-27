package foundation

import (
	"encoding/json"

	supportcontract "github.com/goravel/framework/contracts/support"
)

type Json struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

func NewJson() supportcontract.Json {
	return &Json{
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}
}

func (j *Json) Marshal(v any) ([]byte, error) {
	return j.marshal(v)
}

func (j *Json) Unmarshal(data []byte, v any) error {
	return j.unmarshal(data, v)
}
