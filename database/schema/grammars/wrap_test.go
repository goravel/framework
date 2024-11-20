package grammars

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
)

type WrapTestSuite struct {
	suite.Suite
	wrap *Wrap
}

func TestWrapSuite(t *testing.T) {
	suite.Run(t, new(WrapTestSuite))
}

func (s *WrapTestSuite) SetupTest() {
	s.wrap = NewWrap(database.DriverPostgres, "prefix_")
}

func (s *WrapTestSuite) TestColumnWithAlias() {
	result := s.wrap.Column("column as alias")
	s.Equal(`"column" as "prefix_alias"`, result)
}

func (s *WrapTestSuite) TestColumnWithoutAlias() {
	result := s.wrap.Column("column")
	s.Equal(`"column"`, result)
}

func (s *WrapTestSuite) TestColumnsWithMultipleColumns() {
	result := s.wrap.Columnize([]string{"column1", "column2 as alias2"})
	s.Equal(`"column1", "column2" as "prefix_alias2"`, result)
}

func (s *WrapTestSuite) TestQuoteWithNonEmptyValue() {
	result := s.wrap.Quote("value")
	s.Equal("'value'", result)
}

func (s *WrapTestSuite) TestQuoteWithEmptyValue() {
	result := s.wrap.Quote("")
	s.Equal("", result)
}

func (s *WrapTestSuite) TestQuotes() {
	result := s.wrap.Quotes([]string{"value1", "value2"})
	s.Equal([]string{"'value1'", "'value2'"}, result)

	s.wrap.driver = database.DriverSqlserver
	result = s.wrap.Quotes([]string{"value1", "value2"})
	s.Equal([]string{"N'value1'", "N'value2'"}, result)
}

func (s *WrapTestSuite) TestSegmentsWithMultipleSegments() {
	result := s.wrap.Segments([]string{"table", "column"})
	s.Equal(`"prefix_table"."column"`, result)
}

func (s *WrapTestSuite) TestTableWithAlias() {
	result := s.wrap.Table("table as alias")
	s.Equal(`"prefix_table" as "prefix_alias"`, result)
}

func (s *WrapTestSuite) TestTableWithoutAlias() {
	result := s.wrap.Table("table")
	s.Equal(`"prefix_table"`, result)
}

func (s *WrapTestSuite) TestValueWithAsterisk() {
	result := s.wrap.Value("*")
	s.Equal("*", result)
}

func (s *WrapTestSuite) TestValueWithNonAsterisk() {
	result := s.wrap.Value("value")
	s.Equal(`"value"`, result)
}

func (s *WrapTestSuite) TestValueOfMysql() {
	s.wrap.driver = database.DriverMysql
	result := s.wrap.Value("value")
	s.Equal("`value`", result)
}
