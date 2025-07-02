package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksroute "github.com/goravel/framework/mocks/route"
	mockssession "github.com/goravel/framework/mocks/session"
)

type TestRequestSuite struct {
	suite.Suite
	testRequest        *TestRequest
	mockRoute          *mocksroute.Route
	mockSessionManager *mockssession.Manager
}

func TestTestRequestSuite(t *testing.T) {
	suite.Run(t, new(TestRequestSuite))
}

// SetupTest will run before each test in the suite.
func (s *TestRequestSuite) SetupTest() {
	s.mockRoute = mocksroute.NewRoute(s.T())
	s.mockSessionManager = mockssession.NewManager(s.T())

	s.testRequest = &TestRequest{
		t:                 s.T(),
		ctx:               context.Background(),
		defaultHeaders:    make(map[string]string),
		defaultCookies:    make([]*http.Cookie, 0),
		sessionAttributes: make(map[string]any),
		json:              json.New(),
		route:             s.mockRoute,
		session:           s.mockSessionManager,
	}
}

func (s *TestRequestSuite) TestBindAndCall() {
	s.Run("succeed to bind and call", func() {
		s.mockRoute.EXPECT().Test(httptest.NewRequest("GET", "/", nil).WithContext(context.Background())).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"name": "John", "age": 30}`)),
		}, nil).Once()

		var user struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		response, err := s.testRequest.Bind(&user).Get("/")

		s.NoError(err)
		s.NotNil(response)
		s.Equal("John", user.Name)
		s.Equal(30, user.Age)
	})

	s.Run("should not bind when response is not successful", func() {
		s.mockRoute.EXPECT().Test(httptest.NewRequest("GET", "/", nil).WithContext(context.Background())).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
		}, nil).Once()

		response, err := s.testRequest.Get("/")

		s.NoError(err)
		s.NotNil(response)
		response.AssertInternalServerError()
	})
}

func (s *TestRequestSuite) TestWithCookie() {
	cookie := &http.Cookie{Name: "test", Value: "test"}

	request := s.testRequest.WithCookie(cookie)

	s.Equal(request.(*TestRequest).defaultCookies, []*http.Cookie{cookie})
}

func (s *TestRequestSuite) TestWithCookies() {
	cookies := []*http.Cookie{
		{Name: "test", Value: "test"},
		{Name: "test2", Value: "test2"},
	}

	request := s.testRequest.WithCookies(cookies)

	s.Equal(request.(*TestRequest).defaultCookies, cookies)
}

func (s *TestRequestSuite) TestSetSessionErrors() {
	var (
		mockDriver  *mockssession.Driver
		mockSession *mockssession.Session
	)

	type testCase struct {
		name              string
		setup             func()
		expectedError     string
		sessionAttributes map[string]any
	}

	cases := []testCase{
		{
			name: "DriverError",
			setup: func() {
				s.mockSessionManager.On("Driver").Return(nil, errors.New("driver retrieval error")).Once()
			},
			expectedError:     "driver retrieval error",
			sessionAttributes: map[string]any{"user_id": 123},
		},
		{
			name: "BuildSessionError",
			setup: func() {
				s.mockSessionManager.On("Driver").Return(mockDriver, nil).Once()
				s.mockSessionManager.On("BuildSession", mockDriver).Return(nil, errors.New("build session error")).Once()
			},
			expectedError:     "build session error",
			sessionAttributes: map[string]any{"user_id": 123},
		},
		{
			name: "SaveError",
			setup: func() {
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
			mockDriver = mockssession.NewDriver(s.T())
			mockSession = mockssession.NewSession(s.T())

			tc.setup()

			request := s.testRequest.WithSession(tc.sessionAttributes)

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

	request := s.testRequest.WithSession(sessionAttributes)

	err := request.(*TestRequest).setSession()

	s.NoError(err)
}

func (s *TestRequestSuite) TestSetSessionUsingWithoutSession() {
	s.NoError(s.testRequest.setSession())

	s.mockSessionManager.AssertNotCalled(s.T(), "Driver")
	s.mockSessionManager.AssertNotCalled(s.T(), "BuildSession", mockssession.NewDriver(s.T()))
}
