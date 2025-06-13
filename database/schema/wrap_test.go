package schema

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type WrapTestSuite struct {
	suite.Suite
	wrap *Wrap
}

func TestWrapSuite(t *testing.T) {
	suite.Run(t, new(WrapTestSuite))
}

func (s *WrapTestSuite) SetupTest() {
	s.wrap = NewWrap("prefix_")
	s.wrap.SetValueWrapper(nil)
}

func (s *WrapTestSuite) TestColumn() {
	// With alias
	result := s.wrap.Column("column as alias")
	s.Equal(`"column" as "prefix_alias"`, result)

	// Without alias
	result = s.wrap.Column("column")
	s.Equal(`"column"`, result)
}

func (s *WrapTestSuite) TestColumnize() {
	result := s.wrap.Columnize([]string{"column1", "column2 as alias2"})
	s.Equal(`"column1", "column2" as "prefix_alias2"`, result)
}

func (s *WrapTestSuite) TestJsonFieldAndPath() {
	// without path
	field, path := s.wrap.JsonFieldAndPath("column")
	s.Equal(`"column"`, field)
	s.Equal(``, path)

	// with single path
	field, path = s.wrap.JsonFieldAndPath("column->path")
	s.Equal(`"column"`, field)
	s.Equal(`, '$."path"'`, path)

	// with multiple nested paths
	field, path = s.wrap.JsonFieldAndPath("column->path->nested")
	s.Equal(`"column"`, field)
	s.Equal(`, '$."path"."nested"'`, path)

	// with array index
	field, path = s.wrap.JsonFieldAndPath("column->path[0]")
	s.Equal(`"column"`, field)
	s.Equal(`, '$."path"[0]'`, path)

	// with array index and nested keys
	field, path = s.wrap.JsonFieldAndPath("column->path[0]->nested")
	s.Equal(`"column"`, field)
	s.Equal(`, '$."path"[0]."nested"'`, path)

}

func (s *WrapTestSuite) TestJsonPathAttributes() {
	spiltPath := func(v string) []string {
		return strings.Split(v, "->")
	}
	// with single attribute
	result := s.wrap.JsonPathAttributes(spiltPath("column->path"))
	s.Equal([]string{"'column'", "'path'"}, result)

	// with multiple attributes
	result = s.wrap.JsonPathAttributes(spiltPath("column->path->nested"))
	s.Equal([]string{"'column'", "'path'", "'nested'"}, result)

	// with array index
	result = s.wrap.JsonPathAttributes(spiltPath("column->path[0]"))
	s.Equal([]string{"'column'", "'path'", "0"}, result)

	// with array index and nested keys
	result = s.wrap.JsonPathAttributes(spiltPath("column->path[0]->nested"))
	s.Equal([]string{"'column'", "'path'", "0", "'nested'"}, result)
}

func (s *WrapTestSuite) TestQuote() {
	// With non-empty value
	result := s.wrap.Quote("value")
	s.Equal("'value'", result)

	// With empty value
	result = s.wrap.Quote("")
	s.Equal("", result)
}

func (s *WrapTestSuite) TestQuotes() {
	result := s.wrap.Quotes([]string{"value1", "value2"})
	s.Equal([]string{"'value1'", "'value2'"}, result)
}

func (s *WrapTestSuite) TestSegmentsWithMultipleSegments() {
	result := s.wrap.Segments([]string{"table", "column"})
	s.Equal(`"prefix_table"."column"`, result)
}

func (s *WrapTestSuite) TestTable() {
	// With alias
	result := s.wrap.Table("table as alias")
	s.Equal(`"prefix_table" as "prefix_alias"`, result)

	// With schema
	result = s.wrap.Table("goravel.table")
	s.Equal(`"goravel"."prefix_table"`, result)

	// Without alias
	result = s.wrap.Table("table")
	s.Equal(`"prefix_table"`, result)
}

func (s *WrapTestSuite) TestValue() {
	// With asterisk
	result := s.wrap.Value("*")
	s.Equal("*", result)

	// Without asterisk
	result = s.wrap.Value("value")
	s.Equal(`"value"`, result)
}

func (s *WrapTestSuite) TestAliasedTable() {
	result := s.wrap.aliasedTable("users as u")
	s.Equal(`"prefix_users" as "prefix_u"`, result)
}

func (s *WrapTestSuite) TestAliasedValue() {
	result := s.wrap.aliasedValue("users.name as user_name")
	s.Equal(`"prefix_users"."name" as "prefix_user_name"`, result)
}

func (s *WrapTestSuite) TestSetValueWrapper() {
	// mysql wrapper
	s.wrap.SetValueWrapper(func(value string) string {
		return "`" + strings.ReplaceAll(value, "`", "``") + "`"
	})

	s.Equal("`column` as `prefix_alias`", s.wrap.Column("column as alias"))
	s.Equal("`column1`, `column2` as `prefix_alias2`", s.wrap.Columnize([]string{"column1", "column2 as alias2"}))
	s.Equal("`column`", s.wrap.Value("column"))
	s.Equal("`prefix_table`", s.wrap.Table("table"))
	s.Equal("`prefix_table`.`column`", s.wrap.Segments([]string{"table", "column"}))

	// sql server wrapper
	s.wrap.SetValueWrapper(func(value string) string {
		return "[" + strings.ReplaceAll(value, "]", "]]") + "]"
	})

	s.Equal("[column] as [prefix_alias]", s.wrap.Column("column as alias"))
	s.Equal("[column1], [column2] as [prefix_alias2]", s.wrap.Columnize([]string{"column1", "column2 as alias2"}))
	s.Equal("[column]", s.wrap.Value("column"))
	s.Equal("[prefix_table]", s.wrap.Table("table"))
	s.Equal("[prefix_table].[column]", s.wrap.Segments([]string{"table", "column"}))
}
