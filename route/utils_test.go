package route

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mockshttp "github.com/goravel/framework/mocks/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHTTPHandlerFuncToHandlerFunc(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	goravel := HTTPHandlerFuncToHandlerFunc(handler)

	mockCtx := &mockshttp.Context{}
	mockResponse := &mockshttp.ContextResponse{}
	mockRequest := &mockshttp.ContextRequest{}

	mockCtx.On("Response").Return(mockResponse)
	mockCtx.On("Request").Return(mockRequest)

	mockWriter := httptest.NewRecorder()
	mockResponse.On("Writer").Return(mockWriter)

	mockHttpReq := httptest.NewRequest("GET", "/test", nil)
	mockRequest.On("Origin").Return(mockHttpReq)

	assert.Nil(t, goravel.ServeHTTP(mockCtx))

	assert.True(t, called)
	mockCtx.AssertExpectations(t)
	mockResponse.AssertExpectations(t)
	mockRequest.AssertExpectations(t)
}

func TestHTTPHandlerToHandler(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusCreated)
	})

	goravel := HTTPHandlerToHandler(handler)

	mockCtx := &mockshttp.Context{}
	mockResponse := &mockshttp.ContextResponse{}
	mockRequest := &mockshttp.ContextRequest{}

	mockCtx.On("Response").Return(mockResponse)
	mockCtx.On("Request").Return(mockRequest)

	mockWriter := httptest.NewRecorder()
	mockResponse.On("Writer").Return(mockWriter)

	mockHttpReq := httptest.NewRequest("POST", "/test", nil)
	mockRequest.On("Origin").Return(mockHttpReq)

	assert.Nil(t, goravel.ServeHTTP(mockCtx))

	assert.True(t, called)
	assert.Equal(t, http.StatusCreated, mockWriter.Code)
	mockCtx.AssertExpectations(t)
	mockResponse.AssertExpectations(t)
	mockRequest.AssertExpectations(t)
}

func TestHTTPMiddlewareToMiddleware(t *testing.T) {
	middlewareCalled := false
	handlerCalled := false

	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			w.Header().Set("X-Test", "middleware")
			next.ServeHTTP(w, r)
		})
	}

	goravelMiddleware := HTTPMiddlewareToMiddleware(middleware)

	nextHandler := &mockshttp.Handler{}
	nextHandler.On("ServeHTTP", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		handlerCalled = true
	})

	handler := goravelMiddleware(nextHandler)

	mockCtx := &mockshttp.Context{}
	mockResponse := &mockshttp.ContextResponse{}
	mockRequest := &mockshttp.ContextRequest{}

	mockCtx.On("Response").Return(mockResponse)
	mockCtx.On("Request").Return(mockRequest)

	mockWriter := httptest.NewRecorder()
	mockResponse.On("Writer").Return(mockWriter)

	mockHttpReq := httptest.NewRequest("GET", "/test", nil)
	mockRequest.On("Origin").Return(mockHttpReq)

	assert.Nil(t, handler.ServeHTTP(mockCtx))

	assert.True(t, middlewareCalled)
	assert.True(t, handlerCalled)
	assert.Equal(t, "middleware", mockWriter.Header().Get("X-Test"))
	mockCtx.AssertExpectations(t)
	mockResponse.AssertExpectations(t)
	mockRequest.AssertExpectations(t)
	nextHandler.AssertExpectations(t)
}
