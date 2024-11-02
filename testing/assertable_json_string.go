package testing

import (
	"encoding/json"

	contractstesting "github.com/goravel/framework/contracts/testing"
)

type AssertableJSONString struct {
	json    string
	decoded map[string]any
}

func NewAssertableJSONString(jsonStr string) (contractstesting.AssertableJSON, error) {
	var decoded map[string]any
	err := json.Unmarshal([]byte(jsonStr), &decoded)
	if err != nil {
		return nil, err
	}

	return &AssertableJSONString{
		json:    jsonStr,
		decoded: decoded,
	}, nil
}

func (r *AssertableJSONString) Json() map[string]any {
	return r.decoded
}
