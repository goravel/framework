package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMysqls(t *testing.T) {
	mysql := Mysql()
	assert.NotNil(t, mysql)
	assert.Len(t, containers[ContainerTypeMysql], 1)

	mysqls := Mysqls(2)
	assert.Len(t, mysqls, 2)
	assert.Len(t, containers[ContainerTypeMysql], 2)

	assert.Nil(t, Stop())
}
