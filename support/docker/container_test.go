package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractstesting "github.com/goravel/framework/contracts/testing"
)

func TestContainer(t *testing.T) {
	container := NewContainer()
	assert.NoError(t, container.Remove())

	config1 := contractstesting.DatabaseConfig{
		Host:        "localhost",
		Port:        5432,
		Database:    "goravel",
		Username:    "goravel",
		Password:    "Framework!123",
		ContainerID: "123456",
	}
	config2 := contractstesting.DatabaseConfig{
		Host:        "localhost1",
		Port:        5432,
		Database:    "goravel",
		Username:    "goravel",
		Password:    "Framework!123",
		ContainerID: "123456",
	}

	container.Add(ContainerTypePostgres, config1)
	container.Add(ContainerTypePostgres, config2)
	container.Add(ContainerTypeSqlite, config1)

	containers := container.All()
	assert.Len(t, containers, 2)
	assert.Len(t, containers[ContainerTypePostgres], 2)
	assert.Len(t, containers[ContainerTypeSqlite], 1)
	assert.Equal(t, containers[ContainerTypePostgres][0], config1)
	assert.Equal(t, containers[ContainerTypePostgres][1], config2)
	assert.Equal(t, containers[ContainerTypeSqlite][0], config1)
}
