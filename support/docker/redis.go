package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/goravel/framework/contracts/testing"
)

type Redis struct {
	port        int
	containerID string
	image       *testing.Image
}

func NewRedis() (*Redis, error) {
	return &Redis{
		image: &testing.Image{
			Repository:   "redis",
			Tag:          "latest",
			ExposedPorts: []string{"6379"},
		},
	}, nil
}

func (receiver *Redis) Build() error {
	command, exposedPorts := imageToCommand(receiver.image)
	containerID, err := Run(command)
	if err != nil {
		return fmt.Errorf("init Redis docker error: %v", err)
	}
	if containerID == "" {
		return fmt.Errorf("no container id return when creating Redis docker")
	}

	receiver.containerID = containerID
	receiver.port = getPort(exposedPorts, 6379)

	if err := receiver.connect(); err != nil {
		return fmt.Errorf("connect Redis docker error: %v", err)
	}

	return nil
}

func (receiver *Redis) Config() RedisConfig {
	return RedisConfig{
		Port: receiver.port,
	}
}

func (receiver *Redis) Stop() error {
	if _, err := Run(fmt.Sprintf("docker stop %s", receiver.containerID)); err != nil {
		return fmt.Errorf("stop Redis docker error: %v", err)
	}

	return nil
}

func (receiver *Redis) connect() error {
	var err error
	for i := 0; i < 60; i++ {
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("localhost:%d", receiver.port),
			Password: "",
			DB:       0,
		})

		if _, err = client.Ping(context.Background()).Result(); err == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	return err
}

type RedisConfig struct {
	Port int
}
