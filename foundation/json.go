package foundation

import (
	"encoding/json"

	supportcontract "github.com/goravel/framework/contracts/support"
)

type JSON struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

func NewJSON() supportcontract.Json {
	return &JSON{
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}
}

func (j *JSON) Marshal(v any) ([]byte, error) {
	return j.marshal(v)
}

func (j *JSON) Unmarshal(data []byte, v any) error {
	return j.unmarshal(data, v)
}
