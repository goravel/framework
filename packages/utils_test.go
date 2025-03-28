package packages

import (
	"bytes"
	"go/token"
	"strconv"
	"testing"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/stretchr/testify/suite"
)

type UtilsTestSuite struct {
	suite.Suite
}

func (s *UtilsTestSuite) SetupTest() {}

func (s *UtilsTestSuite) TearDownTest() {}

func TestUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (s *UtilsTestSuite) TestExprExists() {
	s.NotPanics(func() {
		s.Run("expr exists", func() {
			s.True(
				ExprExists(
					MustParseStatement("&some.Struct{}").(dst.Expr),
					MustParseStatement("[]any{&some.Struct{}}").(*dst.CompositeLit).Elts,
				),
			)
		})
		s.Run("expr does not exist", func() {
			s.False(
				ExprExists(
					MustParseStatement("&some.Struct{}").(dst.Expr),
					MustParseStatement("[]any{&some.OtherStruct{}}").(*dst.CompositeLit).Elts,
				),
			)
		})
	})

}

func (s *UtilsTestSuite) TestKeyExists() {
	s.NotPanics(func() {
		s.Run("key exists", func() {
			s.True(
				KeyExists(
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
					MustParseStatement(`map[string]any{"someKey":"exist"}`).(*dst.CompositeLit).Elts,
				),
			)
		})
		s.Run("key does not exist", func() {
			s.False(
				KeyExists(
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
					MustParseStatement(`map[string]any{"otherKey":"exist"}`).(*dst.CompositeLit).Elts,
				),
			)
		})
	})
}

func (s *UtilsTestSuite) TestMustParseStatement() {
	s.Run("parse failed", func() {
		s.Panics(func() {
			MustParseStatement("var invalid:=syntax")
		})
	})

	s.Run("parse success", func() {
		s.NotPanics(func() {
			s.NotNil(MustParseStatement(`struct{x *int}`))
		})
	})
}

func (s *UtilsTestSuite) TestWrapNewline() {
	src := `package main

var value = 1
var _ = map[string]any{"key": &value, "func": func() bool { return true }}
`

	df, err := decorator.Parse(src)
	s.NoError(err)

	// without WrapNewline
	var buf bytes.Buffer
	s.NoError(decorator.Fprint(&buf, df))
	s.Equal(src, buf.String())

	// with WrapNewline
	WrapNewline(df)
	buf.Reset()
	s.NoError(decorator.Fprint(&buf, df))
	s.NotEqual(src, buf.String())
	s.Equal(`package main

var value = 1
var _ = map[string]any{
	"key": &value,
	"func": func() bool {
		return true
	},
}
`, buf.String())

}
