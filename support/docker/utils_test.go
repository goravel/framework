package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	contractstesting "github.com/goravel/framework/contracts/testing"
)

func TestExposedPort(t *testing.T) {
	assert.Equal(t, 1, ExposedPort([]string{"1:2"}, 2))
}

func TestImageToCommand(t *testing.T) {
	command, exposedPorts := ImageToCommand(nil)
	assert.Equal(t, "", command)
	assert.Nil(t, exposedPorts)

	command, exposedPorts = ImageToCommand(&contractstesting.Image{
		Repository: "redis",
		Tag:        "latest",
	})

	assert.Equal(t, "docker run --rm -d redis:latest", command)
	assert.Nil(t, exposedPorts)

	command, exposedPorts = ImageToCommand(&contractstesting.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"6379"},
		Env:          []string{"a=b"},
	})
	assert.Equal(t, fmt.Sprintf("docker run --rm -d -e a=b -p %d:6379 redis:latest", ExposedPort(exposedPorts, 6379)), command)
	assert.True(t, ExposedPort(exposedPorts, 6379) > 0)

	command, exposedPorts = ImageToCommand(&contractstesting.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"1234:6379"},
		Env:          []string{"a=b"},
		Args:         []string{"--a=b"},
	})
	assert.Equal(t, "docker run --rm -d -e a=b -p 1234:6379 redis:latest --a=b", command)
	assert.Equal(t, []string{"1234:6379"}, exposedPorts)
}
