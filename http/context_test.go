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
	var myKey struct{}
	s.ctx.WithValue("Hello", "world")
	s.ctx.WithValue(myKey, "hola")
	s.ctx.WithValue(1, "hi")
	s.ctx.WithValue(2.2, "hey")
	s.Equal("world", s.ctx.Value("Hello"))
	s.Equal("hola", s.ctx.Value(myKey))
	s.Equal("hi", s.ctx.Value(1))
	s.Equal("hey", s.ctx.Value(2.2))
}

func (s *ContextTestSuite) TestRequest() {
	s.Nil(s.ctx.Request())
}

func (s *ContextTestSuite) TestResponse() {
	s.Nil(s.ctx.Response())
}
