package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/mocks/config"
	routemock "github.com/goravel/framework/mocks/route"
)

type HTTPRunnerTestSuite struct {
	suite.Suite
	mockConfig *configmock.Config
	mockRoute  *routemock.Route
	runner     *HTTPRunner
}

func TestHTTPRunnerTestSuite(t *testing.T) {
	suite.Run(t, &HTTPRunnerTestSuite{})
}

func (s *HTTPRunnerTestSuite) SetupTest() {
	s.mockConfig = configmock.NewConfig(s.T())
	s.mockRoute = routemock.NewRoute(s.T())
	s.runner = NewHTTPRunner(s.mockConfig, s.mockRoute)
}

func (s *HTTPRunnerTestSuite) TestNewHTTPRunner() {
	s.NotNil(s.runner)
	s.Equal(s.mockConfig, s.runner.config)
	s.Equal(s.mockRoute, s.runner.route)
}

func (s *HTTPRunnerTestSuite) TestSignature() {
	s.Equal("goravel:http", s.runner.Signature())
}

func (s *HTTPRunnerTestSuite) TestShouldRun_WhenRouteNotNilAndDefaultConfigSet() {
	s.mockConfig.EXPECT().GetString("http.default").Return("gin").Once()
	s.mockConfig.EXPECT().GetBool("app.auto_run", true).Return(true).Once()

	result := s.runner.ShouldRun()

	s.True(result)
}

func (s *HTTPRunnerTestSuite) TestShouldRun_WhenRouteIsNil() {
	// Pass nil directly as route.Route interface to avoid typed nil issue
	runner := NewHTTPRunner(s.mockConfig, nil)

	result := runner.ShouldRun()

	s.False(result)
}

func (s *HTTPRunnerTestSuite) TestShouldRun_WhenDefaultConfigEmpty() {
	s.mockConfig.EXPECT().GetString("http.default").Return("").Once()

	result := s.runner.ShouldRun()

	s.False(result)
}

func (s *HTTPRunnerTestSuite) TestRun_SuccessfullyRunHTTPServerOnly() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()
	s.mockRoute.EXPECT().Run().Return(nil).Once()

	err := s.runner.Run()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestRun_SuccessfullyRunHTTPServerOnly_PortsAreTheSame() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("3000").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()
	s.mockRoute.EXPECT().Run().Return(nil).Once()

	err := s.runner.Run()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestRun_SuccessfullyRunBothHTTPAndHTTPSServers() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("3001").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("/path/to/cert").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("/path/to/key").Once()
	s.mockRoute.EXPECT().RunTLS().Return(nil).Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()
	s.mockRoute.EXPECT().Run().Return(nil).Once()

	err := s.runner.Run()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestRun_ErrorWhenHTTPServerFailsToStart() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()
	s.mockRoute.EXPECT().Run().Return(assert.AnError).Once()

	err := s.runner.Run()

	s.Error(err)
	s.Equal(assert.AnError, err)
}

func (s *HTTPRunnerTestSuite) TestRun_ErrorWhenHTTPSServerFailsToStart() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("3001").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("/path/to/cert").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("/path/to/key").Once()
	s.mockRoute.EXPECT().RunTLS().Return(assert.AnError).Once()

	err := s.runner.Run()

	s.Error(err)
	s.Equal(assert.AnError, err)
}

func (s *HTTPRunnerTestSuite) TestRun_SkipHTTPServerWhenHostIsEmpty() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()

	err := s.runner.Run()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestRun_SkipHTTPServerWhenPortIsEmpty() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("").Once()

	err := s.runner.Run()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestRun_SkipHTTPSServerWhenTLSHostIsEmpty() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("3001").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("/path/to/cert").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("/path/to/key").Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()
	s.mockRoute.EXPECT().Run().Return(nil).Once()

	err := s.runner.Run()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestRun_SkipHTTPSServerWhenTLSPortIsEmpty() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("/path/to/cert").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("/path/to/key").Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()
	s.mockRoute.EXPECT().Run().Return(nil).Once()

	err := s.runner.Run()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestRun_SkipHTTPServerWhenTLSPortEqualsHTTPPort() {
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("3000").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("/path/to/cert").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("/path/to/key").Once()
	s.mockRoute.EXPECT().RunTLS().Return(nil).Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()

	err := s.runner.Run()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestShutdown_Successfully() {
	s.mockRoute.EXPECT().Shutdown().Return(nil).Once()

	err := s.runner.Shutdown()

	s.NoError(err)
}

func (s *HTTPRunnerTestSuite) TestShutdown_ErrorDuringShutdown() {
	s.mockRoute.EXPECT().Shutdown().Return(assert.AnError).Once()

	err := s.runner.Shutdown()

	s.Error(err)
	s.Equal(assert.AnError, err)
}

func (s *HTTPRunnerTestSuite) TestIntegration_FullLifecycle() {
	// Test ShouldRun
	s.mockConfig.EXPECT().GetString("http.default").Return("gin").Once()
	s.mockConfig.EXPECT().GetBool("app.auto_run", true).Return(true).Once()
	s.True(s.runner.ShouldRun())

	// Test Run
	s.mockConfig.EXPECT().GetString("http.tls.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.tls.port").Return("3001").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.cert").Return("/path/to/cert").Once()
	s.mockConfig.EXPECT().GetString("http.tls.ssl.key").Return("/path/to/key").Once()
	s.mockRoute.EXPECT().RunTLS().Return(nil).Once()
	s.mockConfig.EXPECT().GetString("http.host").Return("127.0.0.1").Once()
	s.mockConfig.EXPECT().GetString("http.port").Return("3000").Once()
	s.mockRoute.EXPECT().Run().Return(nil).Once()
	err := s.runner.Run()
	s.NoError(err)

	// Test Shutdown
	s.mockRoute.EXPECT().Shutdown().Return(nil).Once()
	err = s.runner.Shutdown()
	s.NoError(err)
}
