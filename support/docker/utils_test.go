package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/testing/docker"
)

func TestImageToCommand(t *testing.T) {
	command, exposedPorts := ImageToCommand(nil)
	assert.Equal(t, "", command)
	assert.Nil(t, exposedPorts)

	command, exposedPorts = ImageToCommand(&docker.Image{
		Repository: "redis",
		Tag:        "latest",
	})

	assert.Equal(t, "docker run --rm -d redis:latest", command)
	assert.Nil(t, exposedPorts)

	command, exposedPorts = ImageToCommand(&docker.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"6379"},
		Env:          []string{"a=b"},
	})
	assert.Equal(t, fmt.Sprintf("docker run --rm -d -e a=b -p %d:6379 redis:latest", exposedPorts[6379]), command)
	assert.True(t, exposedPorts[6379] > 0)

	command, exposedPorts = ImageToCommand(&docker.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"1234:6379"},
		Env:          []string{"a=b"},
		Args:         []string{"--a=b"},
	})
	assert.Equal(t, "docker run --rm -d -e a=b -p 1234:6379 redis:latest --a=b", command)
	assert.Equal(t, map[int]int{6379: 1234}, exposedPorts)
}
