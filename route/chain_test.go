package route

import (
	"testing"

	"github.com/goravel/framework/contracts/http"
	mockshttp "github.com/goravel/framework/mocks/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChainCreatesMiddlewaresFromSlice(t *testing.T) {
	middleware1 := func(next http.Handler) http.Handler { return next }
	middleware2 := func(next http.Handler) http.Handler { return next }

	middlewares := Chain(middleware1, middleware2)

	assert.Len(t, middlewares, 2)
}

func TestChainHandlerWithNoMiddlewares(t *testing.T) {
	middlewares := Chain()

	finalHandler := &mockshttp.Handler{}
	finalHandler.EXPECT().ServeHTTP(mock.Anything).Return(nil)

	handler := middlewares.Handler(finalHandler)
	mockCtx := &mockshttp.Context{}

	assert.Nil(t, handler.ServeHTTP(mockCtx))

	finalHandler.AssertCalled(t, "ServeHTTP", mockCtx)
}

func TestChainHandlerWithSingleMiddleware(t *testing.T) {
	middlewareCalled := false

	middleware := func(next http.Handler) http.Handler {
		return &middlewareHandler{
			next: next,
			fn: func() {
				middlewareCalled = true
			},
		}
	}

	middlewares := Chain(middleware)

	finalHandler := &mockshttp.Handler{}
	finalHandler.EXPECT().ServeHTTP(mock.Anything).Return(nil)

	handler := middlewares.Handler(finalHandler)
	mockCtx := &mockshttp.Context{}

	assert.Nil(t, handler.ServeHTTP(mockCtx))

	assert.True(t, middlewareCalled)
	finalHandler.AssertCalled(t, "ServeHTTP", mockCtx)
}

func TestChainHandlerWithMultipleMiddlewares(t *testing.T) {
	var executionOrder []string

	middleware1 := func(next http.Handler) http.Handler {
		return &middlewareHandler{
			next: next,
			fn: func() {
				executionOrder = append(executionOrder, "middleware1")
			},
		}
	}
	middleware2 := func(next http.Handler) http.Handler {
		return &middlewareHandler{
			next: next,
			fn: func() {
				executionOrder = append(executionOrder, "middleware2")
			},
		}
	}
	middleware3 := func(next http.Handler) http.Handler {
		return &middlewareHandler{
			next: next,
			fn: func() {
				executionOrder = append(executionOrder, "middleware3")
			},
		}
	}

	middlewares := Chain(middleware1, middleware2, middleware3)

	finalHandler := &mockshttp.Handler{}
	finalHandler.EXPECT().ServeHTTP(mock.Anything).Return(nil).Run(func(ctx http.Context) {
		executionOrder = append(executionOrder, "finalHandler")
	})

	handler := middlewares.Handler(finalHandler)
	mockCtx := &mockshttp.Context{}

	assert.Nil(t, handler.ServeHTTP(mockCtx))

	assert.Equal(t, []string{"middleware1", "middleware2", "middleware3", "finalHandler"}, executionOrder)
	finalHandler.AssertCalled(t, "ServeHTTP", mockCtx)
}

func TestChainHandlerFuncWithMiddlewares(t *testing.T) {
	middlewareCalled := false
	handlerFuncCalled := false

	middleware := func(next http.Handler) http.Handler {
		return &middlewareHandler{
			next: next,
			fn: func() {
				middlewareCalled = true
			},
		}
	}

	middlewares := Chain(middleware)

	handlerFunc := http.HandlerFunc(func(ctx http.Context) error {
		handlerFuncCalled = true
		return nil
	})

	handler := middlewares.HandlerFunc(handlerFunc)
	mockCtx := &mockshttp.Context{}

	assert.Nil(t, handler.ServeHTTP(mockCtx))

	assert.True(t, middlewareCalled)
	assert.True(t, handlerFuncCalled)
}

func TestChainHandlerImplementsHttpHandler(t *testing.T) {
	middlewares := Chain()
	finalHandler := &mockshttp.Handler{}
	finalHandler.EXPECT().ServeHTTP(mock.Anything).Return(nil)

	handler := middlewares.Handler(finalHandler)

	assert.Implements(t, (*http.Handler)(nil), handler)
}

type middlewareHandler struct {
	next http.Handler
	fn   func()
}

func (m *middlewareHandler) ServeHTTP(ctx http.Context) error {
	m.fn()
	return m.next.ServeHTTP(ctx)
}
