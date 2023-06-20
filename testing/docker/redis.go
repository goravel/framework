package docker

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
)

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
