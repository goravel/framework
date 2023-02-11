package limiter

import (
	"github.com/gookit/color"
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/suite"
	"testing"
)

type LimiterTestSuite struct {
	suite.Suite
}

func TestLimiterTestSuite(t *testing.T) {
	suite.Run(t, new(LimiterTestSuite))
}

func (s *LimiterTestSuite) SetupTest() {

}

func (s *LimiterTestSuite) TestLimiterRouteToKeyString() {
	key := RouteToKeyString("http://localhost:8080/api/")
	s.Equal("http_--localhost_8080-api-", key)
}

func (s *LimiterTestSuite) TestLimiter() {
	if !file.Exists("../.env") {
		color.Redln("No limiter tests run, need create .env based on .env.example, then initialize it")
		return
	}
	initConfig()
	httpCtx := http.Background().(*http.GinContext).Instance()
	key := "127.0.0.1"
	rate, err := CheckRate(httpCtx, key, "1-M")
	s.NoError(err)

	// rate.Reached is false
	s.False(rate.Reached)

	// rate.Reached is true
	rate, err = CheckRate(httpCtx, key, "0-M")
	s.NoError(err)
	s.True(rate.Reached)
}

func initConfig() {
	application := config.NewApplication("../.env")
	application.Add("limiter", map[string]any{
		"store": "memory",
	})

	facades.Config = application
}
