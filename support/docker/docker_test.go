package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/env"
)

func TestMysqls(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	mysql := Mysql()
	assert.NotNil(t, mysql)
	assert.Len(t, containers[ContainerTypeMysql], 1)

	mysqls := Mysqls(2)
	assert.Len(t, mysqls, 2)
	assert.Len(t, containers[ContainerTypeMysql], 2)

	assert.Nil(t, Stop())
}
