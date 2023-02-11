package limiter

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/http"
	"github.com/goravel/framework/testing/mock"
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
	mockConfig := mock.Config()
	mockConfig.On("GetString", "limiter.store", "memory").Return("memory")

	ginCtx := http.Background().(*http.GinContext).Instance()

	key := "test1"
	rate, err := CheckRate(ginCtx, key, "1-M")
	s.NoError(err)
	s.False(rate.Reached)

	key = "test2"
	rate, err = CheckRate(ginCtx, key, "0-M")
	s.NoError(err)
	s.True(rate.Reached)
}
