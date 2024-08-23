package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMysql(t *testing.T) {
	configs, err := Mysql1()
	assert.Nil(t, err)
	assert.Len(t, configs, 1)
	assert.Len(t, containers[ContainerTypeMysql], 1)

	configs, err = Mysql1(2)
	assert.Nil(t, err)
	assert.Len(t, configs, 2)
	assert.Len(t, containers[ContainerTypeMysql], 2)

	configs, err = Mysql1(1)
	assert.Nil(t, err)
	assert.Len(t, configs, 1)
	assert.Len(t, containers[ContainerTypeMysql], 2)
}
