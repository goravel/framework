package testing

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	mockssession "github.com/goravel/framework/mocks/session"
)

type TestRequestSuite struct {
	suite.Suite
	mockSessionManager *mockssession.Manager
}

func TestTestRequestSuite(t *testing.T) {
	suite.Run(t, new(TestRequestSuite))
}

// SetupTest will run before each test in the suite.
func (s *TestRequestSuite) SetupTest() {
	s.mockSessionManager = mockssession.NewManager(s.T())
	sessionFacade = s.mockSessionManager
}

func (s *TestRequestSuite) TearDownTest() {
	sessionFacade = nil
}

func (s *TestRequestSuite) TestSetSessionErrors() {
	type testCase struct {
		name              string
		mockBehavior      func(mockDriver *mockssession.Driver, mockSession *mockssession.Session)
		expectedError     string
		sessionAttributes map[string]any
	}

	cases := []testCase{
		{
			name: "DriverError",
			mockBehavior: func(mockDriver *mockssession.Driver, mockSession *mockssession.Session) {
				s.mockSessionManager.On("Driver").Return(nil, errors.New("driver retrieval error")).Once()
			},
			expectedError:     "driver retrieval error",
			sessionAttributes: map[string]any{"user_id": 123},
		},
		{
			name: "BuildSessionError",
			mockBehavior: func(mockDriver *mockssession.Driver, mockSession *mockssession.Session) {
				s.mockSessionManager.On("Driver").Return(mockDriver, nil).Once()
				s.mockSessionManager.On("BuildSession", mockDriver).Return(nil, errors.New("build session error")).Once()
			},
			expectedError:     "build session error",
			sessionAttributes: map[string]any{"user_id": 123},
		},
		{
			name: "SaveError",
			mockBehavior: func(mockDriver *mockssession.Driver, mockSession *mockssession.Session) {
				s.mockSessionManager.On("Driver").Return(mockDriver, nil).Once()
				s.mockSessionManager.On("BuildSession", mockDriver).Return(mockSession, nil).Once()

				mockSession.On("Put", "user_id", 123).Return(mockSession).Once()

				mockSession.On("GetName").Return("session_name").Once()
				mockSession.On("GetID").Return("session_id").Once()

				mockSession.On("Save").Return(errors.New("session save error")).Once()
			},
			expectedError:     "session save error",
			sessionAttributes: map[string]any{"user_id": 123},
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			mockDriver := mockssession.NewDriver(s.T())
			mockSession := mockssession.NewSession(s.T())

			tc.mockBehavior(mockDriver, mockSession)

			request := NewTestRequest(s.T()).WithSession(tc.sessionAttributes)

			err := request.(*TestRequest).setSession()

			if tc.expectedError == "" {
				s.NoError(err)
			} else {
				s.EqualError(err, tc.expectedError)
			}
		})
	}
}

func (s *TestRequestSuite) TestSetSessionUsingWithSession() {
	mockDriver := mockssession.NewDriver(s.T())
	mockSession := mockssession.NewSession(s.T())

	sessionAttributes := map[string]any{
		"user_id":   123,
		"user_role": "admin",
	}

	s.mockSessionManager.On("Driver").Return(mockDriver, nil).Once()

	s.mockSessionManager.On("BuildSession", mockDriver).Return(mockSession, nil).Once()

	for key, value := range sessionAttributes {
		mockSession.On("Put", key, value).Return(mockSession).Once()
	}

	mockSession.On("GetName").Return("session_name").Once()
	mockSession.On("GetID").Return("session_id").Once()

	mockSession.On("Save").Return(nil).Once()
	s.mockSessionManager.On("ReleaseSession", mockSession).Once()

	request := NewTestRequest(s.T()).WithSession(sessionAttributes)

	err := request.(*TestRequest).setSession()

	s.NoError(err)
}

func (s *TestRequestSuite) TestSetSessionUsingWithoutSession() {
	request := NewTestRequest(s.T())

	err := request.(*TestRequest).setSession()

	s.NoError(err)

	s.mockSessionManager.AssertNotCalled(s.T(), "Driver")
	s.mockSessionManager.AssertNotCalled(s.T(), "BuildSession", mockssession.NewDriver(s.T()))
}
