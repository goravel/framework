package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/convert"
)

type ColumnDefinitionTestSuite struct {
	suite.Suite
	columnDefinition *ColumnDefinition
}

func TestColumnDefinitionTestSuite(t *testing.T) {
	suite.Run(t, &ColumnDefinitionTestSuite{})
}

func (s *ColumnDefinitionTestSuite) SetupTest() {
	s.columnDefinition = &ColumnDefinition{}
}

func (s *ColumnDefinitionTestSuite) GetAutoIncrement() {
	s.False(s.columnDefinition.GetAutoIncrement())

	s.columnDefinition.AutoIncrement()
	s.True(s.columnDefinition.GetAutoIncrement())
}

func (s *ColumnDefinitionTestSuite) GetDefault() {
	s.Nil(s.columnDefinition.GetDefault())

	s.columnDefinition.def = "default"
	s.Equal("default", s.columnDefinition.GetDefault())
}

func (s *ColumnDefinitionTestSuite) GetName() {
	s.Empty(s.columnDefinition.GetName())

	s.columnDefinition.name = convert.Pointer("name")
	s.Equal("name", s.columnDefinition.GetName())
}

func (s *ColumnDefinitionTestSuite) GetLength() {
	s.Equal(0, s.columnDefinition.GetLength())

	s.columnDefinition.length = convert.Pointer(255)
	s.Equal(255, s.columnDefinition.GetLength())
}

func (s *ColumnDefinitionTestSuite) GetNullable() {
	s.False(s.columnDefinition.GetNullable())

	s.columnDefinition.nullable = convert.Pointer(true)
	s.True(s.columnDefinition.GetNullable())
}

func (s *ColumnDefinitionTestSuite) GetType() {
	s.Empty(s.columnDefinition.GetType())

	s.columnDefinition.ttype = convert.Pointer("string")
	s.Equal("string", s.columnDefinition.GetType())
}

func (s *ColumnDefinitionTestSuite) Unsigned() {
	s.columnDefinition.Unsigned()
	s.True(*s.columnDefinition.unsigned)
}
