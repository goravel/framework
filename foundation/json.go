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
	if Json == nil {
		Json = &JSON{
			marshal:   json.Marshal,
			unmarshal: json.Unmarshal,
		}
	}

	return Json
}

func (j *JSON) Marshal(v any) ([]byte, error) {
	return j.marshal(v)
}

func (j *JSON) MarshalString(v any) (string, error) {
	s, err := j.marshal(v)
	return string(s), err
}

func (j *JSON) Unmarshal(data []byte, v any) error {
	return j.unmarshal(data, v)
}

func (j *JSON) UnmarshalString(data string, v any) error {
	return j.unmarshal([]byte(data), v)
}
