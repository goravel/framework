package foundation

import (
	"encoding/json"

	foundationcontract "github.com/goravel/framework/contracts/foundation"
)

type Json struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

func NewJson() foundationcontract.Json {
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
