package grammars

import (
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
}

func (s *WrapTestSuite) ColumnWithAlias() {
	result := s.wrap.Column("column as alias")
	s.Equal(`"column" as "prefix_alias"`, result)
}

func (s *WrapTestSuite) ColumnWithoutAlias() {
	result := s.wrap.Column("column")
	s.Equal(`"column"`, result)
}

func (s *WrapTestSuite) ColumnsWithMultipleColumns() {
	result := s.wrap.Columnize([]string{"column1", "column2 as alias2"})
	s.Equal(`"column1", "column2" as "prefix_alias2"`, result)
}

func (s *WrapTestSuite) QuoteWithNonEmptyValue() {
	result := s.wrap.Quote("value")
	s.Equal("'value'", result)
}

func (s *WrapTestSuite) QuoteWithEmptyValue() {
	result := s.wrap.Quote("")
	s.Equal("", result)
}

func (s *WrapTestSuite) SegmentsWithMultipleSegments() {
	result := s.wrap.Segments([]string{"table", "column"})
	s.Equal(`"prefix_table"."column"`, result)
}

func (s *WrapTestSuite) TableWithAlias() {
	result := s.wrap.Table("table as alias")
	s.Equal(`"prefix_table" as "prefix_alias"`, result)
}

func (s *WrapTestSuite) TableWithoutAlias() {
	result := s.wrap.Table("table")
	s.Equal(`"prefix_table"`, result)
}

func (s *WrapTestSuite) ValueWithAsterisk() {
	result := s.wrap.Value("*")
	s.Equal("*", result)
}

func (s *WrapTestSuite) ValueWithNonAsterisk() {
	result := s.wrap.Value("value")
	s.Equal(`"value"`, result)
}
