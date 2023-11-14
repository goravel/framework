package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/env"
)

func TestGetExposedPort(t *testing.T) {
	assert.Equal(t, 1, getExposedPort([]string{"1:2"}, 2))
}

func TestGetValidPort(t *testing.T) {
	assert.True(t, getValidPort() > 0)
}

func TestImageToCommand(t *testing.T) {
	command, exposedPorts := imageToCommand(nil)
	assert.Equal(t, "", command)
	assert.Nil(t, exposedPorts)

	command, exposedPorts = imageToCommand(&contractstesting.Image{
		Repository: "redis",
		Tag:        "latest",
	})

	assert.Equal(t, "docker run --rm -d redis:latest", command)
	assert.Nil(t, exposedPorts)

	command, exposedPorts = imageToCommand(&contractstesting.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"6379"},
		Env:          []string{"a=b"},
	})
	assert.Equal(t, fmt.Sprintf("docker run --rm -d -e a=b -p %d:6379 redis:latest", getExposedPort(exposedPorts, 6379)), command)
	assert.True(t, getExposedPort(exposedPorts, 6379) > 0)

	command, exposedPorts = imageToCommand(&contractstesting.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"1234:6379"},
		Env:          []string{"a=b"},
	})
	assert.Equal(t, "docker run --rm -d -e a=b -p 1234:6379 redis:latest", command)
	assert.Equal(t, []string{"1234:6379"}, exposedPorts)
}

func TestRun(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	_, err := run("ls")
	assert.Nil(t, err)
}
