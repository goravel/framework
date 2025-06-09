package docker

import (
	"fmt"

	contractsdocker "github.com/goravel/framework/contracts/testing/docker"
	"github.com/goravel/framework/errors"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/process"
)

type ImageDriver struct {
	config contractsdocker.ImageConfig
	image  contractsdocker.Image
}

func NewImageDriver(image contractsdocker.Image) *ImageDriver {
	return &ImageDriver{
		image: image,
	}
}

func (r *ImageDriver) Build() error {
	command, exposedPorts := supportdocker.ImageToCommand(&r.image)
	containerID, err := process.Run(command)
	if err != nil {
		return errors.TestingImageBuildFailed.Args(r.image.Repository, err)
	}
	if containerID == "" {
		return errors.TestingImageNoContainerId.Args(r.image.Repository)
	}

	r.config = contractsdocker.ImageConfig{
		ContainerID:  containerID,
		ExposedPorts: exposedPorts,
	}

	return nil
}

func (r *ImageDriver) Config() contractsdocker.ImageConfig {
	return r.config
}

func (r *ImageDriver) Shutdown() error {
	if r.config.ContainerID != "" {
		if _, err := process.Run(fmt.Sprintf("docker stop %s", r.config.ContainerID)); err != nil {
			return errors.TestingImageStopFailed.Args(r.image.Repository, err)
		}
	}

	return nil
}
