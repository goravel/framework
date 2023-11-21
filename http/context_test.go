package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
	ctx *Context
}

func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

func (s *ContextTestSuite) SetupTest() {
	s.ctx = NewContext()
}

func (s *ContextTestSuite) TestContext() {
	s.Equal(context.Background(), s.ctx.Context())
}

func (s *ContextTestSuite) TestWithValue() {
	s.ctx.WithValue("Hello", "world")
	s.Equal("world", s.ctx.Value("Hello"))
}

func (s *ContextTestSuite) TestRequest() {
	s.Nil(s.ctx.Request())
}

func (s *ContextTestSuite) TestResponse() {
	s.Nil(s.ctx.Response())
}
