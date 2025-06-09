package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/testing/docker"
)

func TestExposedPort(t *testing.T) {
	assert.Equal(t, "1", ExposedPort([]string{"1:2"}, "2"))
	assert.Equal(t, "1", ExposedPort([]string{"1:2/udp"}, "2"))
}

func TestImageToCommand(t *testing.T) {
	command, exposedPorts := ImageToCommand(nil)
	assert.Equal(t, "", command)
	assert.Nil(t, exposedPorts)

	command, exposedPorts = ImageToCommand(&docker.Image{
		Repository: "redis",
		Tag:        "latest",
	})

	assert.Equal(t, "docker run --rm -d redis:latest", command)
	assert.True(t, len(exposedPorts) == 0)

	command, exposedPorts = ImageToCommand(&docker.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"6379"},
		Env:          []string{"a=b"},
	})
	assert.Equal(t, fmt.Sprintf("docker run --rm -d -e a=b -p %s:6379 redis:latest", ExposedPort(exposedPorts, "6379")), command)
	assert.NotEmpty(t, ExposedPort(exposedPorts, "6379"))

	command, exposedPorts = ImageToCommand(&docker.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"1234:6379"},
		Env:          []string{"a=b"},
		Args:         []string{"--a=b"},
	})
	assert.Equal(t, "docker run --rm -d -e a=b -p 1234:6379 redis:latest --a=b", command)
	assert.Equal(t, []string{"1234:6379"}, exposedPorts)

	command, _ = ImageToCommand(&docker.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"1234:6379"},
		Env:          []string{"a=b"},
		Args:         []string{"--a=b"},
		Cmd:          []string{"sleep", "1000"},
	})
	assert.Equal(t, "docker run --rm -d -e a=b -p 1234:6379 redis:latest --a=b sleep 1000", command)
}
