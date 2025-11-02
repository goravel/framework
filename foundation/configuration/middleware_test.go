package configuration

import (
	"fmt"
	"reflect"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/stretchr/testify/suite"
)

// makeMiddleware returns a middleware function annotated with an id via closure.
func makeMiddleware(id string) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		fmt.Println(id)
	}
}

type MiddlewareTestSuite struct {
	suite.Suite
	middleware *Middleware
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

func (s *MiddlewareTestSuite) SetupTest() {
	s.middleware = NewMiddleware([]contractshttp.Middleware{makeMiddleware("a"), makeMiddleware("b")})
}

func (s *MiddlewareTestSuite) TestAppend() {
	// Initial length
	s.Equal(2, len(s.middleware.GetGlobalMiddleware()))
	s.middleware.Append(makeMiddleware("c"))
	s.Equal(3, len(s.middleware.GetGlobalMiddleware()))
	s.middleware.Append(makeMiddleware("d"), makeMiddleware("e"))
	s.Equal(5, len(s.middleware.GetGlobalMiddleware()))
	// Append empty does nothing
	s.middleware.Append()
	s.Equal(5, len(s.middleware.GetGlobalMiddleware()))
}

func (s *MiddlewareTestSuite) TestPrepend() {
	s.Equal(2, len(s.middleware.GetGlobalMiddleware()))
	s.middleware.Prepend(makeMiddleware("z"))
	s.Equal(3, len(s.middleware.GetGlobalMiddleware()))
	s.middleware.Prepend(makeMiddleware("x"), makeMiddleware("y"))
	s.Equal(5, len(s.middleware.GetGlobalMiddleware()))
	// Prepend empty does nothing
	s.middleware.Prepend()
	s.Equal(5, len(s.middleware.GetGlobalMiddleware()))
}

func (s *MiddlewareTestSuite) TestUse() {
	s.middleware.Use(makeMiddleware("n"), makeMiddleware("m"))
	s.Equal(2, len(s.middleware.GetGlobalMiddleware()))
	// Use with empty slice resets to empty
	s.middleware.Use()
	s.Empty(s.middleware.GetGlobalMiddleware())
}

func (s *MiddlewareTestSuite) TestRecover() {
	// default recover should be nil
	s.Nil(s.middleware.GetRecover())

	called := false
	fn := func(ctx contractshttp.Context, err any) { called = true }
	s.middleware.Recover(fn)
	s.NotNil(s.middleware.GetRecover())
	s.Equal(reflect.ValueOf(fn).Pointer(), reflect.ValueOf(s.middleware.GetRecover()).Pointer())

	// invoke recover function to ensure it executes
	s.middleware.GetRecover()(nil, nil)
	s.True(called)
}
