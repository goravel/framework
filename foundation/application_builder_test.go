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
	s.Run("Without any With functions", func() {
		s.mockApp.EXPECT().Boot().Return().Once()

		app := s.builder.Create()

		s.NotNil(app)
	})

	s.Run("WithConfig", func() {
		calledConfig := false

		s.mockApp.EXPECT().Boot().Return().Once()

		app := s.builder.
			WithConfig(func() {
				calledConfig = true
			}).
			Create()

		s.NotNil(app)
		s.True(calledConfig)
	})

	s.Run("WithEvents but Event facade is nil", func() {
		s.mockApp.EXPECT().Boot().Return().Once()
		s.mockApp.EXPECT().MakeEvent().Return(nil).Once()

		app := s.builder.
			WithEvents(map[event.Event][]event.Listener{
				mocksevent.NewEvent(s.T()): {mocksevent.NewListener(s.T())},
			}).
			Create()

		s.NotNil(app)
	})

	s.Run("WithEvents but Events is empty", func() {
		s.mockApp.EXPECT().Boot().Return().Once()

		app := s.builder.
			WithEvents(map[event.Event][]event.Listener{}).
			Create()

		s.NotNil(app)
	})

	s.Run("WithEvents", func() {
		s.mockApp.EXPECT().Boot().Return().Once()
		mockEvent := mocksevent.NewEvent(s.T())
		mockListener := mocksevent.NewListener(s.T())
		mockEventFacade := mocksevent.NewInstance(s.T())
		mockEventFacade.EXPECT().
			Register(map[event.Event][]event.Listener{
				mockEvent: {mockListener},
			}).
			Return().
			Once()
		s.mockApp.EXPECT().MakeEvent().Return(mockEventFacade).Once()

		app := s.builder.
			WithEvents(map[event.Event][]event.Listener{
				mockEvent: {mockListener},
			}).
			Create()

		s.NotNil(app)
	})
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
