package docker

import (
	"context"

	"github.com/go-redis/redis/v8"
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

func Redis() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := Pool()
	if err != nil {
		return nil, nil, err
	}
	resource, err := Resource(pool, &dockertest.RunOptions{
		Repository: "redis",
		Tag:        "latest",
		Env:        []string{},
	})
	if err != nil {
		return nil, nil, err
	}
	_ = resource.Expire(600)

	if err := pool.Retry(func() error {
		client := redis.NewClient(&redis.Options{
			Addr:     "localhost:" + resource.GetPort("6379/tcp"),
			Password: "",
			DB:       0,
		})

		if _, err := client.Ping(context.Background()).Result(); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return pool, resource, nil
}
