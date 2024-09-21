package grammars

import (
	"testing"

	"github.com/stretchr/testify/suite"

	schemacontract "github.com/goravel/framework/contracts/database/migration"
	mockschema "github.com/goravel/framework/mocks/database/migration"
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

func (s *PostgresSuite) TestCompileDropIfExists() {
	mockBlueprint := &mockschema.Blueprint{}
	mockBlueprint.On("GetTableName").Return("users").Once()

	s.Equal("drop table if exists users", s.grammar.CompileDropIfExists(mockBlueprint))
}

func (s *PostgresSuite) TestModifyDefault() {
	var (
		mockBlueprint *mockschema.Blueprint
		mockColumn    *mockschema.ColumnDefinition
	)

	tests := []struct {
		name      string
		setup     func()
		expectSql string
	}{
		{
			name: "with change and AutoIncrement",
			setup: func() {
				mockColumn.EXPECT().GetChange().Return(true).Once()
				mockColumn.EXPECT().GetAutoIncrement().Return(true).Once()
			},
		},
		{
			name: "with change and not AutoIncrement, default is nil",
			setup: func() {
				mockColumn.EXPECT().GetChange().Return(true).Once()
				mockColumn.EXPECT().GetAutoIncrement().Return(false).Once()
				mockColumn.EXPECT().GetDefault().Return(nil).Once()
			},
			expectSql: "drop default",
		},
		{
			name: "with change and not AutoIncrement, default is not nil",
			setup: func() {
				mockColumn.EXPECT().GetChange().Return(true).Once()
				mockColumn.EXPECT().GetAutoIncrement().Return(false).Once()
				mockColumn.EXPECT().GetDefault().Return("goravel").Twice()
			},
			expectSql: "set default 'goravel'",
		},
		{
			name: "without change and default is nil",
			setup: func() {
				mockColumn.EXPECT().GetChange().Return(false).Once()
				mockColumn.EXPECT().GetDefault().Return(nil).Once()
			},
		},
		{
			name: "without change and default is not nil",
			setup: func() {
				mockColumn.EXPECT().GetChange().Return(false).Once()
				mockColumn.EXPECT().GetDefault().Return("goravel").Twice()
			},
			expectSql: " default 'goravel'",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockBlueprint = &mockschema.Blueprint{}
			mockColumn = &mockschema.ColumnDefinition{}

			test.setup()

			sql := s.grammar.ModifyDefault(mockBlueprint, mockColumn)

			s.Equal(test.expectSql, sql)

			mockBlueprint.AssertExpectations(s.T())
			mockColumn.AssertExpectations(s.T())
		})
	}
}

func (s *PostgresSuite) TestModifyNullable() {
	mockBlueprint := mockschema.NewBlueprint(s.T())

	mockColumn := mockschema.NewColumnDefinition(s.T())
	mockColumn.EXPECT().GetChange().Return(true).Once()
	mockColumn.EXPECT().GetNullable().Return(true).Once()

	s.Equal("drop not null", s.grammar.ModifyNullable(mockBlueprint, mockColumn))

	mockColumn.EXPECT().GetChange().Return(true).Once()
	mockColumn.EXPECT().GetNullable().Return(false).Once()

	s.Equal("set not null", s.grammar.ModifyNullable(mockBlueprint, mockColumn))

	mockColumn.EXPECT().GetChange().Return(false).Once()
	mockColumn.EXPECT().GetNullable().Return(true).Once()

	s.Equal(" null", s.grammar.ModifyNullable(mockBlueprint, mockColumn))

	mockColumn.EXPECT().GetChange().Return(false).Once()
	mockColumn.EXPECT().GetNullable().Return(false).Once()

	s.Equal(" not null", s.grammar.ModifyNullable(mockBlueprint, mockColumn))
}

func (s *PostgresSuite) TestModifyIncrement() {
	mockBlueprint := mockschema.NewBlueprint(s.T())

	mockColumn := mockschema.NewColumnDefinition(s.T())
	mockColumn.EXPECT().GetChange().Return(true).Once()

	s.Empty(s.grammar.ModifyIncrement(mockBlueprint, mockColumn))

	mockColumn.EXPECT().GetChange().Return(false).Once()
	mockBlueprint.EXPECT().HasCommand("primary").Return(false).Once()
	mockColumn.EXPECT().GetType().Return("bigInteger").Once()
	mockColumn.EXPECT().GetAutoIncrement().Return(true).Once()

	s.Equal(" primary key", s.grammar.ModifyIncrement(mockBlueprint, mockColumn))
}

func (s *PostgresSuite) TestTypeBigInteger() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetAutoIncrement").Return(true).Once()

	s.Equal("bigserial", s.grammar.TypeBigInteger(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetAutoIncrement").Return(false).Once()

	s.Equal("bigint", s.grammar.TypeBigInteger(mockColumn2))
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
