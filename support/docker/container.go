package docker

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support/file"
)

type Container struct {
	file string
}

func NewContainer() *Container {
	return &Container{
		file: filepath.Join(os.TempDir(), "goravel_docker.txt"),
	}
}

func (r *Container) Add(containerType ContainerType, config testing.DatabaseConfig) {
	containerTypeToContainers := r.All()
	containerTypeToContainers[containerType] = append(containerTypeToContainers[containerType], config)

	f, err := os.OpenFile(r.file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := json.NewJson().Marshal(containerTypeToContainers)
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(string(content))
	if err != nil {
		panic(err)
	}
}

func (r *Container) All() map[ContainerType][]testing.DatabaseConfig {
	var containerTypeToContainers map[ContainerType][]testing.DatabaseConfig
	if !file.Exists(r.file) {
		return make(map[ContainerType][]testing.DatabaseConfig)
	}

	f, err := os.OpenFile(r.file, os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := json.NewJson().Unmarshal(bytes.TrimSpace(content), &containerTypeToContainers); err != nil {
		panic(err)
	}

	return containerTypeToContainers
}

func (r *Container) Remove() error {
	return file.Remove(r.file)
}
