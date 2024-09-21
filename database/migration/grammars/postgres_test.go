package grammars

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractsmigration "github.com/goravel/framework/contracts/database/migration"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
)

type PostgresSuite struct {
	suite.Suite
	grammar *Postgres
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, &PostgresSuite{})
}

func (s *PostgresSuite) SetupTest() {
	postgres := &Postgres{
		attributeCommands: []string{"comment"},
		serials:           []string{"bigInteger"},
	}
	postgres.modifiers = []func(contractsmigration.Blueprint, contractsmigration.ColumnDefinition) string{
		postgres.ModifyDefault,
	}

	s.grammar = postgres
}

func (s *PostgresSuite) TestCompileCreate() {
	mockColumn1 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn1.EXPECT().GetName().Return("id").Once()
	mockColumn1.EXPECT().GetType().Return("integer").Once()
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()
	mockColumn1.EXPECT().GetChange().Return(false).Once()
	mockColumn1.EXPECT().GetDefault().Return(nil).Once()

	mockColumn2 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn2.EXPECT().GetName().Return("name").Once()
	mockColumn2.EXPECT().GetType().Return("string").Once()
	mockColumn2.EXPECT().GetLength().Return(100).Once()
	mockColumn2.EXPECT().GetChange().Return(false).Once()
	mockColumn2.EXPECT().GetDefault().Return(nil).Once()

	mockBlueprint := mocksmigration.NewBlueprint(s.T())
	mockBlueprint.EXPECT().GetTableName().Return("users").Once()
	mockBlueprint.EXPECT().GetAddedColumns().Return([]contractsmigration.ColumnDefinition{
		mockColumn1, mockColumn2,
	}).Once()

	s.Equal("create table users (id serial,name varchar(100))",
		s.grammar.CompileCreate(mockBlueprint, nil))
}

func (s *PostgresSuite) TestCompileDropIfExists() {
	mockBlueprint := mocksmigration.NewBlueprint(s.T())
	mockBlueprint.EXPECT().GetTableName().Return("users").Once()

	s.Equal("drop table if exists users", s.grammar.CompileDropIfExists(mockBlueprint))
}

func (s *PostgresSuite) TestModifyDefault() {
	var (
		mockBlueprint *mocksmigration.Blueprint
		mockColumn    *mocksmigration.ColumnDefinition
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
			mockBlueprint = mocksmigration.NewBlueprint(s.T())
			mockColumn = mocksmigration.NewColumnDefinition(s.T())

			test.setup()

			sql := s.grammar.ModifyDefault(mockBlueprint, mockColumn)

			s.Equal(test.expectSql, sql)
		})
	}
}

func (s *PostgresSuite) TestModifyNullable() {
	mockBlueprint := mocksmigration.NewBlueprint(s.T())

	mockColumn := mocksmigration.NewColumnDefinition(s.T())
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
	mockBlueprint := mocksmigration.NewBlueprint(s.T())

	mockColumn := mocksmigration.NewColumnDefinition(s.T())
	mockColumn.EXPECT().GetChange().Return(true).Once()

	s.Empty(s.grammar.ModifyIncrement(mockBlueprint, mockColumn))

	mockColumn.EXPECT().GetChange().Return(false).Once()
	mockBlueprint.EXPECT().HasCommand("primary").Return(false).Once()
	mockColumn.EXPECT().GetType().Return("bigInteger").Once()
	mockColumn.EXPECT().GetAutoIncrement().Return(true).Once()

	s.Equal(" primary key", s.grammar.ModifyIncrement(mockBlueprint, mockColumn))
}

func (s *PostgresSuite) TestTypeBigInteger() {
	mockColumn1 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()

	s.Equal("bigserial", s.grammar.TypeBigInteger(mockColumn1))

	mockColumn2 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn2.EXPECT().GetAutoIncrement().Return(false).Once()

	s.Equal("bigint", s.grammar.TypeBigInteger(mockColumn2))
}

func (s *PostgresSuite) TestTypeInteger() {
	mockColumn1 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()

	s.Equal("serial", s.grammar.TypeInteger(mockColumn1))

	mockColumn2 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn2.EXPECT().GetAutoIncrement().Return(false).Once()

	s.Equal("integer", s.grammar.TypeInteger(mockColumn2))
}

func (s *PostgresSuite) TestTypeString() {
	mockColumn1 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn1.EXPECT().GetLength().Return(100).Once()

	s.Equal("varchar(100)", s.grammar.TypeString(mockColumn1))

	mockColumn2 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn2.EXPECT().GetLength().Return(0).Once()

	s.Equal("varchar", s.grammar.TypeString(mockColumn2))
}
