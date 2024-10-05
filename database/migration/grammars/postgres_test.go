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
	s.grammar = NewPostgres()
}

func (s *PostgresSuite) TestCompileChange() {
	var (
		mockBlueprint *mocksmigration.Blueprint
		mockColumn1   *mocksmigration.ColumnDefinition
		mockColumn2   *mocksmigration.ColumnDefinition
	)

	tests := []struct {
		name      string
		setup     func()
		expectSql string
	}{
		{
			name: "no changes",
			setup: func() {
				mockBlueprint.EXPECT().GetChangedColumns().Return([]contractsmigration.ColumnDefinition{}).Once()
			},
		},
		{
			name: "single change",
			setup: func() {
				mockColumn1.EXPECT().GetAutoIncrement().Return(false).Once()
				mockColumn1.EXPECT().GetDefault().Return("goravel").Twice()
				mockColumn1.EXPECT().GetName().Return("name").Once()
				mockColumn1.EXPECT().GetChange().Return(true).Times(3)
				mockColumn1.EXPECT().GetNullable().Return(true).Once()
				mockBlueprint.EXPECT().GetTableName().Return("users").Once()
				mockBlueprint.EXPECT().GetChangedColumns().Return([]contractsmigration.ColumnDefinition{mockColumn1}).Once()
			},
			expectSql: "alter table users alter column name set default 'goravel', alter column name drop not null",
		},
		{
			name: "multiple changes",
			setup: func() {
				mockColumn1.EXPECT().GetAutoIncrement().Return(false).Once()
				mockColumn1.EXPECT().GetDefault().Return("goravel").Twice()
				mockColumn1.EXPECT().GetName().Return("name").Once()
				mockColumn1.EXPECT().GetChange().Return(true).Times(3)
				mockColumn1.EXPECT().GetNullable().Return(true).Once()
				mockColumn2.EXPECT().GetAutoIncrement().Return(false).Once()
				mockColumn2.EXPECT().GetDefault().Return(1).Twice()
				mockColumn2.EXPECT().GetName().Return("age").Once()
				mockColumn2.EXPECT().GetChange().Return(true).Times(3)
				mockColumn2.EXPECT().GetNullable().Return(false).Once()
				mockBlueprint.EXPECT().GetTableName().Return("users").Once()
				mockBlueprint.EXPECT().GetChangedColumns().Return([]contractsmigration.ColumnDefinition{mockColumn1, mockColumn2}).Once()
			},
			expectSql: "alter table users alter column name set default 'goravel', alter column name drop not null, alter column age set default '1', alter column age set not null",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockBlueprint = mocksmigration.NewBlueprint(s.T())
			mockColumn1 = mocksmigration.NewColumnDefinition(s.T())
			mockColumn2 = mocksmigration.NewColumnDefinition(s.T())

			test.setup()

			sql := s.grammar.CompileChange(mockBlueprint)

			s.Equal(test.expectSql, sql)
		})
	}
}

func (s *PostgresSuite) TestCompileCreate() {
	mockColumn1 := mocksmigration.NewColumnDefinition(s.T())
	mockColumn2 := mocksmigration.NewColumnDefinition(s.T())
	mockBlueprint := mocksmigration.NewBlueprint(s.T())

	// postgres.go::CompileCreate
	mockBlueprint.EXPECT().GetTableName().Return("users").Once()
	// utils.go::getColumns
	mockBlueprint.EXPECT().GetAddedColumns().Return([]contractsmigration.ColumnDefinition{
		mockColumn1, mockColumn2,
	}).Once()
	// utils.go::getColumns
	mockColumn1.EXPECT().GetName().Return("id").Once()
	// utils.go::getType
	mockColumn1.EXPECT().GetType().Return("integer").Once()
	// postgres.go::TypeInteger
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()
	// postgres.go::ModifyDefault
	mockColumn1.EXPECT().GetChange().Return(false).Once()
	mockColumn1.EXPECT().GetDefault().Return(nil).Once()
	// postgres.go::ModifyIncrement
	mockColumn1.EXPECT().GetChange().Return(false).Once()
	mockBlueprint.EXPECT().HasCommand("primary").Return(false).Once()
	mockColumn1.EXPECT().GetType().Return("integer").Once()
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()
	// postgres.go::ModifyNullable
	mockColumn1.EXPECT().GetChange().Return(false).Once()
	mockColumn1.EXPECT().GetNullable().Return(false).Once()

	// utils.go::getColumns
	mockColumn2.EXPECT().GetName().Return("name").Once()
	// utils.go::getType
	mockColumn2.EXPECT().GetType().Return("string").Once()
	// postgres.go::TypeString
	mockColumn2.EXPECT().GetLength().Return(100).Once()
	// postgres.go::ModifyDefault
	mockColumn2.EXPECT().GetChange().Return(false).Once()
	mockColumn2.EXPECT().GetDefault().Return(nil).Once()
	// postgres.go::ModifyIncrement
	mockColumn2.EXPECT().GetChange().Return(false).Once()
	mockBlueprint.EXPECT().HasCommand("primary").Return(false).Once()
	mockColumn2.EXPECT().GetType().Return("string").Once()
	// postgres.go::ModifyNullable
	mockColumn2.EXPECT().GetChange().Return(false).Once()
	mockColumn2.EXPECT().GetNullable().Return(true).Once()

	s.Equal("create table users (id serial primary key not null,name varchar(100) null)",
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
