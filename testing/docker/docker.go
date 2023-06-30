package docker

import (
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"
)

func Pool() (*dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, errors.WithMessage(err, "Could not construct pool")
	}

	if err := pool.Client.Ping(); err != nil {
		return nil, errors.WithMessage(err, "Could not connect to Docker")
	}

	return pool, nil
}

func Resource(pool *dockertest.Pool, opts *dockertest.RunOptions) (*dockertest.Resource, error) {
	return pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
}
