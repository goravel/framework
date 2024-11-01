package testing

import (
	"encoding/json"

	contractstesting "github.com/goravel/framework/contracts/testing"
)

type AssertableJSONString struct {
	json    string
	decoded map[string]any
}

func NewAssertableJSONString(jsonAble string) (contractstesting.AssertableJSON, error) {
	var decoded = make(map[string]any)
	err := json.Unmarshal([]byte(jsonAble), &decoded)
	if err != nil {
		return nil, err
	}

	return &AssertableJSONString{
		json:    jsonAble,
		decoded: decoded,
	}, nil
}

func (r *AssertableJSONString) Json() map[string]any {
	return r.decoded
}
