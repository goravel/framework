package grammars

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractsschema "github.com/goravel/framework/contracts/database/schema"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
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

func (s *PostgresSuite) TestCompileAdd() {
	mockBlueprint := mocksschema.NewBlueprint(s.T())
	mockColumn := mocksschema.NewColumnDefinition(s.T())

	mockBlueprint.EXPECT().GetTableName().Return("users").Once()
	mockColumn.EXPECT().GetName().Return("name").Once()
	mockColumn.EXPECT().GetType().Return("string").Twice()
	mockColumn.EXPECT().GetDefault().Return("goravel").Twice()
	mockColumn.EXPECT().GetNullable().Return(false).Once()
	mockColumn.EXPECT().GetLength().Return(1).Once()
	mockBlueprint.EXPECT().HasCommand("primary").Return(false).Once()

	sql := s.grammar.CompileAdd(mockBlueprint, &contractsschema.Command{
		Column: mockColumn,
	})

	s.Equal("alter table users add column name varchar(1) default 'goravel' not null", sql)
}

func (s *PostgresSuite) TestCompileCreate() {
	mockColumn1 := mocksschema.NewColumnDefinition(s.T())
	mockColumn2 := mocksschema.NewColumnDefinition(s.T())
	mockBlueprint := mocksschema.NewBlueprint(s.T())

	// postgres.go::CompileCreate
	mockBlueprint.EXPECT().GetTableName().Return("users").Once()
	// utils.go::getColumns
	mockBlueprint.EXPECT().GetAddedColumns().Return([]contractsschema.ColumnDefinition{
		mockColumn1, mockColumn2,
	}).Once()
	// utils.go::getColumns
	mockColumn1.EXPECT().GetName().Return("id").Once()
	// utils.go::getType
	mockColumn1.EXPECT().GetType().Return("integer").Once()
	// postgres.go::TypeInteger
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()
	// postgres.go::ModifyDefault
	mockColumn1.EXPECT().GetDefault().Return(nil).Once()
	// postgres.go::ModifyIncrement
	mockBlueprint.EXPECT().HasCommand("primary").Return(false).Once()
	mockColumn1.EXPECT().GetType().Return("integer").Once()
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()
	// postgres.go::ModifyNullable
	mockColumn1.EXPECT().GetNullable().Return(false).Once()

	// utils.go::getColumns
	mockColumn2.EXPECT().GetName().Return("name").Once()
	// utils.go::getType
	mockColumn2.EXPECT().GetType().Return("string").Once()
	// postgres.go::TypeString
	mockColumn2.EXPECT().GetLength().Return(100).Once()
	// postgres.go::ModifyDefault
	mockColumn2.EXPECT().GetDefault().Return(nil).Once()
	// postgres.go::ModifyIncrement
	mockBlueprint.EXPECT().HasCommand("primary").Return(false).Once()
	mockColumn2.EXPECT().GetType().Return("string").Once()
	// postgres.go::ModifyNullable
	mockColumn2.EXPECT().GetNullable().Return(true).Once()

	s.Equal("create table users (id serial primary key not null,name varchar(100) null)",
		s.grammar.CompileCreate(mockBlueprint, nil))
}

func (s *PostgresSuite) TestCompileDropIfExists() {
	mockBlueprint := mocksschema.NewBlueprint(s.T())
	mockBlueprint.EXPECT().GetTableName().Return("users").Once()

	s.Equal("drop table if exists users", s.grammar.CompileDropIfExists(mockBlueprint))
}

func (s *PostgresSuite) TestEscapeNames() {
	// SingleName
	names := []string{"username"}
	expected := []string{`"username"`}
	s.Equal(expected, s.grammar.EscapeNames(names))

	// MultipleNames
	names = []string{"username", "user.email"}
	expected = []string{`"username"`, `"user"."email"`}
	s.Equal(expected, s.grammar.EscapeNames(names))

	// NamesEmpty
	names = []string{}
	expected = []string{}
	s.Equal(expected, s.grammar.EscapeNames(names))
}

func (s *PostgresSuite) TestModifyDefault() {
	var (
		mockBlueprint *mocksschema.Blueprint
		mockColumn    *mocksschema.ColumnDefinition
	)

	tests := []struct {
		name      string
		setup     func()
		expectSql string
	}{
		{
			name: "without change and default is nil",
			setup: func() {
				mockColumn.EXPECT().GetDefault().Return(nil).Once()
			},
		},
		{
			name: "without change and default is not nil",
			setup: func() {
				mockColumn.EXPECT().GetDefault().Return("goravel").Twice()
			},
			expectSql: " default 'goravel'",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockBlueprint = mocksschema.NewBlueprint(s.T())
			mockColumn = mocksschema.NewColumnDefinition(s.T())

			test.setup()

			sql := s.grammar.ModifyDefault(mockBlueprint, mockColumn)

			s.Equal(test.expectSql, sql)
		})
	}
}

func (s *PostgresSuite) TestModifyNullable() {
	mockBlueprint := mocksschema.NewBlueprint(s.T())

	mockColumn := mocksschema.NewColumnDefinition(s.T())

	mockColumn.EXPECT().GetNullable().Return(true).Once()

	s.Equal(" null", s.grammar.ModifyNullable(mockBlueprint, mockColumn))

	mockColumn.EXPECT().GetNullable().Return(false).Once()

	s.Equal(" not null", s.grammar.ModifyNullable(mockBlueprint, mockColumn))
}

func (s *PostgresSuite) TestModifyIncrement() {
	mockBlueprint := mocksschema.NewBlueprint(s.T())

	mockColumn := mocksschema.NewColumnDefinition(s.T())
	mockBlueprint.EXPECT().HasCommand("primary").Return(false).Once()
	mockColumn.EXPECT().GetType().Return("bigInteger").Once()
	mockColumn.EXPECT().GetAutoIncrement().Return(true).Once()

	s.Equal(" primary key", s.grammar.ModifyIncrement(mockBlueprint, mockColumn))
}

func (s *PostgresSuite) TestTypeBigInteger() {
	mockColumn1 := mocksschema.NewColumnDefinition(s.T())
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()

	s.Equal("bigserial", s.grammar.TypeBigInteger(mockColumn1))

	mockColumn2 := mocksschema.NewColumnDefinition(s.T())
	mockColumn2.EXPECT().GetAutoIncrement().Return(false).Once()

	s.Equal("bigint", s.grammar.TypeBigInteger(mockColumn2))
}

func (s *PostgresSuite) TestTypeInteger() {
	mockColumn1 := mocksschema.NewColumnDefinition(s.T())
	mockColumn1.EXPECT().GetAutoIncrement().Return(true).Once()

	s.Equal("serial", s.grammar.TypeInteger(mockColumn1))

	mockColumn2 := mocksschema.NewColumnDefinition(s.T())
	mockColumn2.EXPECT().GetAutoIncrement().Return(false).Once()

	s.Equal("integer", s.grammar.TypeInteger(mockColumn2))
}

func (s *PostgresSuite) TestTypeString() {
	mockColumn1 := mocksschema.NewColumnDefinition(s.T())
	mockColumn1.EXPECT().GetLength().Return(100).Once()

	s.Equal("varchar(100)", s.grammar.TypeString(mockColumn1))

	mockColumn2 := mocksschema.NewColumnDefinition(s.T())
	mockColumn2.EXPECT().GetLength().Return(0).Once()

	s.Equal("varchar", s.grammar.TypeString(mockColumn2))
}
