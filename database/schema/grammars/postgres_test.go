package grammars

import (
	"testing"

	"github.com/stretchr/testify/suite"

	schemacontract "github.com/goravel/framework/contracts/database/schema"
	mockschema "github.com/goravel/framework/mocks/database/schema"
)

type PostgresSuite struct {
	suite.Suite
	grammar *Postgres
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, &PostgresSuite{})
}

func (s *PostgresSuite) SetupTest() {
	s.grammar = NewPostgres()
}

func (s *PostgresSuite) TestCompileComment() {
	mockBlueprint := &mockschema.Blueprint{}
	mockColumnDefinition := &mockschema.ColumnDefinition{}
	mockBlueprint.On("GetTableName").Return("users").Once()
	mockColumnDefinition.On("GetName").Return("id").Once()
	mockColumnDefinition.On("GetComment").Return("comment").Once()

	sql := s.grammar.CompileComment(mockBlueprint, &schemacontract.Command{
		Column: mockColumnDefinition,
	})

	s.Equal("comment on column users.id is 'comment'", sql)
}

func (s *PostgresSuite) TestCompileCreate() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetName").Return("id").Once()
	mockColumn1.On("GetType").Return("string").Once()
	mockColumn1.On("GetLength").Return(100).Once()
	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetName").Return("name").Once()
	mockColumn2.On("GetType").Return("string").Once()
	mockColumn2.On("GetLength").Return(0).Once()
	mockBlueprint := &mockschema.Blueprint{}
	mockBlueprint.On("GetTableName").Return("users").Once()
	mockBlueprint.On("GetAddedColumns").Return([]schemacontract.ColumnDefinition{
		mockColumn1, mockColumn2,
	}).Once()

	s.Equal("create table users (id varchar(100),name varchar)",
		s.grammar.CompileCreate(mockBlueprint, nil))
}

func (s *PostgresSuite) TestCompileDropColumn() {
	mockBlueprint := &mockschema.Blueprint{}
	mockBlueprint.On("GetTableName").Return("users").Once()

	sql := s.grammar.CompileDropColumn(mockBlueprint, &schemacontract.Command{
		Columns: []string{"id", "name"},
	})

	s.Equal("alter table users drop column id,drop column name", sql)
}

func (s *PostgresSuite) TestTypeBigInteger() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetAutoIncrement").Return(true).Once()

	s.Equal("bigserial", s.grammar.TypeBigInteger(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetAutoIncrement").Return(false).Once()

	s.Equal("bigint", s.grammar.TypeBigInteger(mockColumn2))
}

func (s *PostgresSuite) TestTypeChar() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetLength").Return(100).Once()

	s.Equal("char(100)", s.grammar.TypeChar(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetLength").Return(0).Once()

	s.Equal("char", s.grammar.TypeChar(mockColumn2))
}

func (s *PostgresSuite) TestTypeDecimal() {
	mockColumn := &mockschema.ColumnDefinition{}
	mockColumn.On("GetTotal").Return(4).Once()
	mockColumn.On("GetPlaces").Return(2).Once()

	s.Equal("decimal(4, 2)", s.grammar.TypeDecimal(mockColumn))
}

func (s *PostgresSuite) TestTypeEnum() {
	mockColumn := &mockschema.ColumnDefinition{}
	mockColumn.On("GetName").Return("name").Once()
	mockColumn.On("GetAllowed").Return([]string{"a", "b"}).Once()

	s.Equal(`varchar(255) check ("name" in (a,b))`, s.grammar.TypeEnum(mockColumn))
}

func (s *PostgresSuite) TestTypeInteger() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetAutoIncrement").Return(true).Once()

	s.Equal("serial", s.grammar.TypeInteger(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetAutoIncrement").Return(false).Once()

	s.Equal("integer", s.grammar.TypeInteger(mockColumn2))
}

func (s *PostgresSuite) TestTypeString() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetLength").Return(100).Once()

	s.Equal("varchar(100)", s.grammar.TypeString(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetLength").Return(0).Once()

	s.Equal("varchar", s.grammar.TypeString(mockColumn2))
}

func (s *PostgresSuite) TestTypeTime() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetPrecision").Return(1).Once()

	s.Equal("time(1) without time zone", s.grammar.TypeTime(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetPrecision").Return(0).Once()

	s.Equal("time", s.grammar.TypeTime(mockColumn2))
}

func (s *PostgresSuite) TestTypeTimeTz() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetPrecision").Return(1).Once()

	s.Equal("time(1) with time zone", s.grammar.TypeTimeTz(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetPrecision").Return(0).Once()

	s.Equal("time", s.grammar.TypeTimeTz(mockColumn2))
}

func (s *PostgresSuite) TestTypeTimestamp() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetPrecision").Return(1).Once()

	s.Equal("timestamp(1) without time zone", s.grammar.TypeTimestamp(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetPrecision").Return(0).Once()

	s.Equal("timestamp", s.grammar.TypeTimestamp(mockColumn2))
}

func (s *PostgresSuite) TestTypeTimestampTz() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetPrecision").Return(1).Once()

	s.Equal("timestamp(1) with time zone", s.grammar.TypeTimestampTz(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetPrecision").Return(0).Once()

	s.Equal("timestamp", s.grammar.TypeTimestampTz(mockColumn2))
}

func (s *PostgresSuite) TestGetColumns() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetName").Return("id").Once()
	mockColumn1.On("GetType").Return("string").Once()
	mockColumn1.On("GetLength").Return(100).Once()
	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetName").Return("name").Once()
	mockColumn2.On("GetType").Return("string").Once()
	mockColumn2.On("GetLength").Return(0).Once()
	mockBlueprint := &mockschema.Blueprint{}
	mockBlueprint.On("GetTableName").Return("users").Once()
	mockBlueprint.On("GetAddedColumns").Return([]schemacontract.ColumnDefinition{
		mockColumn1, mockColumn2,
	}).Once()

	s.Equal([]string{"id varchar(100)", "name varchar"}, s.grammar.getColumns(mockBlueprint))
}
