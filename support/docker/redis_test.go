package docker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/env"
)

type RedisTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	redis      *Redis
}

func TestRedisTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, new(RedisTestSuite))
}

func (s *RedisTestSuite) SetupTest() {
	s.mockConfig = &configmocks.Config{}
	s.redis = NewRedis()
}

func (s *RedisTestSuite) TestBuild() {
	ctx := context.Background()

	s.Nil(s.redis.Build())
	instance, err := s.redis.connect()
	s.Nil(err)
	s.NotNil(instance)

	s.True(s.redis.Config().Port > 0)

	s.Nil(instance.Set(ctx, "hello", "goravel", 10*time.Second).Err())
	s.Equal("goravel", instance.Get(ctx, "hello").Val())

	s.Nil(s.redis.Stop())
}
