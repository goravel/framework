package foundation

import (
	"testing"

	"github.com/stretchr/testify/suite"

	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

type ApplicationBuilderTestSuite struct {
	suite.Suite
	builder *ApplicationBuilder
	mockApp *mocksfoundation.Application
}

func TestApplicationBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationBuilderTestSuite))
}

func (s *ApplicationBuilderTestSuite) SetupTest() {
	s.mockApp = mocksfoundation.NewApplication(s.T())
	s.builder = &ApplicationBuilder{
		app: s.mockApp,
	}
}

func (s *ApplicationBuilderTestSuite) TestCreate() {
	s.mockApp.EXPECT().Boot().Return().Once()

	calledConfig := false

	app := s.builder.WithConfig(func() {
		calledConfig = true
	}).Create()

	s.NotNil(app)
	s.True(calledConfig)
}

func (s *ApplicationBuilderTestSuite) TestRun() {
	s.mockApp.EXPECT().Boot().Return().Once()
	s.mockApp.EXPECT().Run().Return().Once()

	s.builder.Run()
}

func (s *ApplicationBuilderTestSuite) TestWithConfig() {
	fn := func() {}

	builder := s.builder.WithConfig(fn)

	s.NotNil(builder)
	s.NotNil(s.builder.config)
}
