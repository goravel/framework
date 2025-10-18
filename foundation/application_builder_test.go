package foundation

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/event"
	mocksevent "github.com/goravel/framework/mocks/event"
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
	mockEvent := mocksevent.NewInstance(s.T())
	mockEvent.EXPECT().Register(map[event.Event][]event.Listener{}).Return().Once()
	s.mockApp.EXPECT().MakeEvent().Return(mockEvent).Once()

	calledConfig := false

	app := s.builder.
		WithConfig(func() {
			calledConfig = true
		}).
		WithEvents(map[event.Event][]event.Listener{}).
		Create()

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
