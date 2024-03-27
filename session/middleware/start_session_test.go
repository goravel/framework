package middleware

import (
	"context"
	"testing"
	"time"

	contractshttp "github.com/goravel/framework/contracts/http"
	configmocks "github.com/goravel/framework/mocks/config"
	httpmocks "github.com/goravel/framework/mocks/http"
	sessionmocks "github.com/goravel/framework/mocks/session"
	"github.com/goravel/framework/session"
)

func TestStartSession(t *testing.T) {
	var (
		ctx               *TestContext
		mockConfig        *configmocks.Config
		mockSessionFacade *sessionmocks.Manager
	)

	tests := []struct {
		name   string
		setup  func()
		assert func()
	}{
		{
			name: "Test StartSession",
			setup: func() {
				ctx.request.On("HasSession").Return(false).Once()
				mockDriver := sessionmocks.NewDriver(t)
				mockSessionFacade.On("Driver").Return(mockDriver, nil).Once()
				mockSession := sessionmocks.NewSession(t)
				mockSessionFacade.On("BuildSession", mockDriver).Return(mockSession).Once()
				mockSession.On("GetName").Return("goravel_session").Twice()
				ctx.request.On("Cookie", "goravel_session").Return("").Once()
				mockSession.On("SetID", "").Return(mockSession).Once()
				mockSession.On("Start").Return(true).Once()
				mockConfig.On("Get", "session.lottery").Return([]int{1, 100}).Once()
				ctx.request.On("SetSession", mockSession).Return(ctx.request).Once()
				ctx.request.On("Next").Return().Once()
				ctx.request.On("Session").Return(mockSession).Once()
				mockSession.On("GetID").Return("123456").Once()

				mockConfig.On("GetInt", "session.lifetime").Return(60).Once()
				mockConfig.On("GetString", "session.path").Return("/").Once()
				mockConfig.On("GetString", "session.domain").Return("").Once()
				mockConfig.On("GetBool", "session.secure").Return(false).Once()
				mockConfig.On("GetBool", "session.http_only").Return(true).Once()
				mockConfig.On("GetString", "session.same_site").Return("").Once()

				ctx.response.On("Cookie", contractshttp.Cookie{
					Name:     "goravel_session",
					Value:    "123456",
					Path:     "/",
					Domain:   "",
					Secure:   false,
					HttpOnly: true,
					SameSite: "",
				}).Return(ctx.response).Once()

				mockSession.On("Save").Return(nil).Once()
			},
			assert: func() {
				StartSession()(ctx)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockRequest := httpmocks.NewContextRequest(t)
			mockResponse := httpmocks.NewContextResponse(t)
			ctx = &TestContext{
				request:  mockRequest,
				response: mockResponse,
			}
			mockConfig = &configmocks.Config{}
			mockSessionFacade = &sessionmocks.Manager{}
			session.ConfigFacade = mockConfig
			session.Facade = mockSessionFacade
			//test.setup()
			//test.assert()

			mockConfig.AssertExpectations(t)
			mockSessionFacade.AssertExpectations(t)
		})
	}
}

type TestContext struct {
	response *httpmocks.ContextResponse
	request  *httpmocks.ContextRequest
}

func (r *TestContext) Deadline() (deadline time.Time, ok bool) {
	panic("do not need to implement it")
}

func (r *TestContext) Done() <-chan struct{} {
	panic("do not need to implement it")
}

func (r *TestContext) Err() error {
	panic("do not need to implement it")
}

func (r *TestContext) Value(any) any {
	panic("do not need to implement it")
}

func (r *TestContext) Context() context.Context {
	panic("do not need to implement it")
}

func (r *TestContext) WithValue(string, any) {
	panic("do not need to implement it")
}

func (r *TestContext) Request() contractshttp.ContextRequest {
	return r.request
}

func (r *TestContext) Response() contractshttp.ContextResponse {
	return r.response
}
