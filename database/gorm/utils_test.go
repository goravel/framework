package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyStruct(t *testing.T) {
	type Data struct {
		Name string
		age  int
	}

	data := copyStruct(Data{Name: "name", age: 18})

	assert.Equal(t, "name", data.Field(0).Interface().(string))
	assert.Panics(t, func() {
		data.Field(1).Interface()
	})
}
