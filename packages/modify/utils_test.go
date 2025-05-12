package modify

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
					MustParseExpr("[]any{&some.Struct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
			s.NotEqual(-1,
				ExprIndex(
					MustParseExpr("[]any{&some.Struct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
		})
		s.Run("expr does not exist", func() {
			s.False(
				ExprExists(
					MustParseExpr("[]any{&some.OtherStruct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
			s.Equal(-1,
				ExprIndex(
					MustParseExpr("[]any{&some.OtherStruct{}}").(*dst.CompositeLit).Elts,
					MustParseExpr("&some.Struct{}").(dst.Expr),
				),
			)
		})
	})

}

func (s *UtilsTestSuite) TestUsesImport() {
	df, err := decorator.Parse(`package main
import (
    "fmt"        
    mylog "log"
)

func main() {
    fmt.Println("hello")
}`)
	s.Require().NoError(err)
	s.Require().NotNil(df)

	s.True(IsUsingImport(df, "fmt"))
	s.False(IsUsingImport(df, "log", "mylog"))
}

func (s *UtilsTestSuite) TestKeyExists() {
	s.NotPanics(func() {
		s.Run("key exists", func() {
			s.True(
				KeyExists(
					MustParseExpr(`map[string]any{"someKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
			s.NotEqual(-1,
				KeyIndex(
					MustParseExpr(`map[string]any{"someKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
		})
		s.Run("key does not exist", func() {
			s.False(
				KeyExists(
					MustParseExpr(`map[string]any{"otherKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
			s.Equal(-1,
				KeyIndex(
					MustParseExpr(`map[string]any{"otherKey":"exist"}`).(*dst.CompositeLit).Elts,
					&dst.BasicLit{Kind: token.STRING, Value: strconv.Quote("someKey")},
				),
			)
		})
	})
}

func (s *UtilsTestSuite) TestMustParseStatement() {
	s.Run("parse failed", func() {
		s.Panics(func() {
			MustParseExpr("var invalid:=syntax")
		})
	})

	s.Run("parse success", func() {
		s.NotPanics(func() {
			s.NotNil(MustParseExpr(`struct{x *int}`))
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
