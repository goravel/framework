package migration

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/migration"
	"github.com/goravel/framework/database/migration/grammars"
	mocksmigration "github.com/goravel/framework/mocks/database/migration"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	"github.com/goravel/framework/support/convert"
)

type BlueprintTestSuite struct {
	suite.Suite
	blueprint *Blueprint
	grammars  map[database.Driver]migration.Grammar
}

func TestBlueprintTestSuite(t *testing.T) {
	suite.Run(t, &BlueprintTestSuite{
		grammars: map[database.Driver]migration.Grammar{
			database.DriverPostgres: grammars.NewPostgres(),
		},
	})
}

func (s *BlueprintTestSuite) SetupTest() {
	s.blueprint = NewBlueprint("goravel_", "users")
}

func (s *BlueprintTestSuite) TestAddAttributeCommands() {
	var (
		mockGrammar      *mocksmigration.Grammar
		columnDefinition = &ColumnDefinition{
			comment: convert.Pointer("comment"),
		}
	)

	tests := []struct {
		name           string
		columns        []*ColumnDefinition
		setup          func()
		expectCommands []*migration.Command
	}{
		{
			name: "Should not add command when columns is empty",
			setup: func() {
				mockGrammar.EXPECT().GetAttributeCommands().Return([]string{"test"}).Once()
			},
		},
		{
			name:    "Should not add command when columns is not empty but GetAttributeCommands does not contain a valid command",
			columns: []*ColumnDefinition{columnDefinition},
			setup: func() {
				mockGrammar.EXPECT().GetAttributeCommands().Return([]string{"test"}).Once()
			},
		},
		{
			name:    "Should add comment command when columns is not empty and GetAttributeCommands contains a comment command",
			columns: []*ColumnDefinition{columnDefinition},
			setup: func() {
				mockGrammar.EXPECT().GetAttributeCommands().Return([]string{"comment"}).Once()
			},
			expectCommands: []*migration.Command{
				{
					Column: columnDefinition,
					Name:   "comment",
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockGrammar = mocksmigration.NewGrammar(s.T())
			s.blueprint.columns = test.columns
			test.setup()

			s.blueprint.addAttributeCommands(mockGrammar)
			s.Equal(test.expectCommands, s.blueprint.commands)
		})
	}
}

func (s *BlueprintTestSuite) TestAddImpliedCommands() {
	var (
		mockGrammar *mocksmigration.Grammar
	)

	tests := []struct {
		name           string
		columns        []*ColumnDefinition
		commands       []*migration.Command
		setup          func()
		expectCommands []*migration.Command
	}{
		{
			name: "Should not add the add command when there are added columns but it is a create operation",
			columns: []*ColumnDefinition{
				{
					name: convert.Pointer("name"),
				},
			},
			commands: []*migration.Command{
				{
					Name: "create",
				},
			},
			setup: func() {
				mockGrammar.EXPECT().GetAttributeCommands().Return([]string{}).Once()
			},
			expectCommands: []*migration.Command{
				{
					Name: "create",
				},
			},
		},
		{
			name: "Should not add the change command when there are changed columns but it is a create operation",
			columns: []*ColumnDefinition{
				{
					name:   convert.Pointer("name"),
					change: convert.Pointer(true),
				},
			},
			commands: []*migration.Command{
				{
					Name: "create",
				},
			},
			setup: func() {
				mockGrammar.EXPECT().GetAttributeCommands().Return([]string{}).Once()
			},
			expectCommands: []*migration.Command{
				{
					Name: "create",
				},
			},
		},
		{
			name: "Should add the add, change, attribute commands when there are added and changed columns, and it is not a create operation",
			columns: []*ColumnDefinition{
				{
					name:    convert.Pointer("name"),
					comment: convert.Pointer("comment"),
				},
				{
					name:   convert.Pointer("age"),
					change: convert.Pointer(true),
				},
			},
			setup: func() {
				mockGrammar.EXPECT().GetAttributeCommands().Return([]string{"comment"}).Once()
			},
			expectCommands: []*migration.Command{
				{
					Name: "add",
				},
				{
					Name: "change",
				},
				{
					Column: &ColumnDefinition{
						name:    convert.Pointer("name"),
						comment: convert.Pointer("comment"),
					},
					Name: "comment",
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockGrammar = mocksmigration.NewGrammar(s.T())
			s.blueprint.columns = test.columns
			s.blueprint.commands = test.commands
			test.setup()

			s.blueprint.addImpliedCommands(mockGrammar)
			s.Equal(test.expectCommands, s.blueprint.commands)
		})
	}
}

func (s *BlueprintTestSuite) TestBigIncrements() {
	name := "name"
	s.blueprint.BigIncrements(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("bigInteger"),
	})
}

func (s *BlueprintTestSuite) TestBigInteger() {
	name := "name"
	s.blueprint.BigInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer("bigInteger"),
	})

	s.blueprint.BigInteger(name).AutoIncrement().Unsigned()
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("bigInteger"),
	})
}

func (s *BlueprintTestSuite) TestBuild() {
	for _, grammar := range s.grammars {
		mockQuery := mocksorm.NewQuery(s.T())

		s.blueprint.Create()
		s.blueprint.String("name")

		sqlStatements := s.blueprint.ToSql(mockQuery, grammar)
		s.NotEmpty(sqlStatements)

		mockQuery.EXPECT().Exec(sqlStatements[0]).Return(nil, nil).Once()
		s.Nil(s.blueprint.Build(mockQuery, grammar))

		sqlStatements = s.blueprint.ToSql(mockQuery, grammar)
		s.NotEmpty(sqlStatements)

		mockQuery.EXPECT().Exec(sqlStatements[0]).Return(nil, errors.New("error")).Once()
		s.EqualError(s.blueprint.Build(mockQuery, grammar), "error")
	}
}

func (s *BlueprintTestSuite) TestGetAddedColumns() {
	name := "name"
	change := true
	addedColumn := &ColumnDefinition{
		name: &name,
	}
	changedColumn := &ColumnDefinition{
		change: &change,
		name:   &name,
	}

	s.blueprint.columns = []*ColumnDefinition{addedColumn, changedColumn}

	s.Len(s.blueprint.GetAddedColumns(), 1)
	s.Equal(addedColumn, s.blueprint.GetAddedColumns()[0])
}

func (s *BlueprintTestSuite) TestGetChangedColumns() {
	name := "name"
	change := true
	addedColumn := &ColumnDefinition{
		name: &name,
	}
	changedColumn := &ColumnDefinition{
		change: &change,
		name:   &name,
	}

	s.blueprint.columns = []*ColumnDefinition{addedColumn, changedColumn}

	s.Len(s.blueprint.GetChangedColumns(), 1)
	s.Equal(changedColumn, s.blueprint.GetChangedColumns()[0])
}

func (s *BlueprintTestSuite) TestGetTableName() {
	s.blueprint.SetTable("users")
	s.Equal("goravel_users", s.blueprint.GetTableName())
}

func (s *BlueprintTestSuite) TestHasCommand() {
	s.False(s.blueprint.HasCommand(commandCreate))
	s.blueprint.Create()
	s.True(s.blueprint.HasCommand(commandCreate))
}

func (s *BlueprintTestSuite) TestIsCreate() {
	s.False(s.blueprint.isCreate())
	s.blueprint.commands = []*migration.Command{
		{
			Name: commandCreate,
		},
	}
	s.True(s.blueprint.isCreate())
}

func (s *BlueprintTestSuite) TestID() {
	s.blueprint.ID()
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          convert.Pointer("id"),
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("bigInteger"),
	})

	s.blueprint.ID("name")
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          convert.Pointer("name"),
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("bigInteger"),
	})
}

func (s *BlueprintTestSuite) TestInteger() {
	name := "name"
	s.blueprint.Integer(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer("integer"),
	})

	s.blueprint.Integer(name).AutoIncrement().Unsigned()
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("integer"),
	})
}

func (s *BlueprintTestSuite) TestString() {
	column := "name"
	customLength := 100
	length := defaultStringLength
	ttype := "string"
	s.blueprint.String(column)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		length: &length,
		name:   &column,
		ttype:  &ttype,
	})

	s.blueprint.String(column, customLength)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		length: &customLength,
		name:   &column,
		ttype:  &ttype,
	})
}

func (s *BlueprintTestSuite) TestToSql() {
	for driver, grammar := range s.grammars {
		mockQuery := mocksorm.NewQuery(s.T())
		s.blueprint.Create()
		s.blueprint.String("name")
		// TODO Add below when implementing the comment method
		//s.blueprint.String("name").Comment("comment")
		//s.blueprint.Comment("comment")

		if driver == database.DriverPostgres {
			s.Len(s.blueprint.ToSql(mockQuery, grammar), 1)
		} else {
			s.Empty(s.blueprint.ToSql(mockQuery, grammar))
		}
	}
}

func (s *BlueprintTestSuite) TestUnsignedBigInteger() {
	name := "name"
	s.blueprint.UnsignedBigInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:     &name,
		ttype:    convert.Pointer("bigInteger"),
		unsigned: convert.Pointer(true),
	})
}
