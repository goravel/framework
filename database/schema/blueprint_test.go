package schema

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/constants"
	"github.com/goravel/framework/database/schema/grammars"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	"github.com/goravel/framework/support/convert"
)

type BlueprintTestSuite struct {
	suite.Suite
	blueprint *Blueprint
	grammars  map[database.Driver]schema.Grammar
}

func TestBlueprintTestSuite(t *testing.T) {
	suite.Run(t, &BlueprintTestSuite{
		grammars: map[database.Driver]schema.Grammar{
			database.DriverPostgres: grammars.NewPostgres("goravel_"),
		},
	})
}

func (s *BlueprintTestSuite) SetupTest() {
	s.blueprint = NewBlueprint("goravel_", "users")
}

func (s *BlueprintTestSuite) TestAddAttributeCommands() {
	var (
		mockGrammar      *mocksschema.Grammar
		columnDefinition = &ColumnDefinition{
			comment: convert.Pointer("comment"),
		}
	)

	tests := []struct {
		name           string
		columns        []*ColumnDefinition
		setup          func()
		expectCommands []*schema.Command
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
			expectCommands: []*schema.Command{
				{
					Column: columnDefinition,
					Name:   "comment",
				},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockGrammar = mocksschema.NewGrammar(s.T())
			s.blueprint.columns = test.columns
			test.setup()

			s.blueprint.addAttributeCommands(mockGrammar)
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

		sqlStatements := s.blueprint.ToSql(grammar)
		s.NotEmpty(sqlStatements)

		mockQuery.EXPECT().Exec(sqlStatements[0]).Return(nil, nil).Once()
		s.Nil(s.blueprint.Build(mockQuery, grammar))

		sqlStatements = s.blueprint.ToSql(grammar)
		s.NotEmpty(sqlStatements)

		mockQuery.EXPECT().Exec(sqlStatements[0]).Return(nil, errors.New("error")).Once()
		s.EqualError(s.blueprint.Build(mockQuery, grammar), "error")
	}
}

func (s *BlueprintTestSuite) TestChar() {
	column := "name"
	customLength := 100
	length := constants.DefaultStringLength
	ttype := "char"
	s.blueprint.Char(column)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		length: &length,
		name:   &column,
		ttype:  &ttype,
	})

	s.blueprint.Char(column, customLength)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		length: &customLength,
		name:   &column,
		ttype:  &ttype,
	})
}

func (s *BlueprintTestSuite) TestCreateIndexName() {
	name := s.blueprint.createIndexName("index", []string{"id", "name-1", "name.2"})
	s.Equal("goravel_users_id_name_1_name_2_index", name)

	s.blueprint.table = "public.users"
	name = s.blueprint.createIndexName("index", []string{"id", "name-1", "name.2"})
	s.Equal("public_goravel_users_id_name_1_name_2_index", name)
}

func (s *BlueprintTestSuite) TestDecimal() {
	name := "name"
	s.blueprint.Decimal(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer("decimal"),
	})
}

func (s *BlueprintTestSuite) TestDouble() {
	name := "name"
	s.blueprint.Double(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer("double"),
	})
}

func (s *BlueprintTestSuite) TestFloat() {
	name := "name"
	s.blueprint.Float(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:      &name,
		precision: convert.Pointer(53),
		ttype:     convert.Pointer("float"),
	})

	s.blueprint.Float(name, 10)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:      &name,
		precision: convert.Pointer(10),
		ttype:     convert.Pointer("float"),
	})
}

func (s *BlueprintTestSuite) TestGetAddedColumns() {
	name := "name"
	addedColumn := &ColumnDefinition{
		name: &name,
	}

	s.blueprint.columns = []*ColumnDefinition{addedColumn}

	s.Len(s.blueprint.GetAddedColumns(), 1)
	s.Equal(addedColumn, s.blueprint.GetAddedColumns()[0])
}

func (s *BlueprintTestSuite) TestHasCommand() {
	s.False(s.blueprint.HasCommand(constants.CommandCreate))
	s.blueprint.Create()
	s.True(s.blueprint.HasCommand(constants.CommandCreate))
}

func (s *BlueprintTestSuite) TestIntegerIncrements() {
	name := "name"
	s.blueprint.IntegerIncrements(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("integer"),
	})
}

func (s *BlueprintTestSuite) TestIndexCommand() {
	s.blueprint.indexCommand("index", []string{"id", "name"})
	s.Contains(s.blueprint.commands, &schema.Command{
		Columns: []string{"id", "name"},
		Name:    "index",
		Index:   "goravel_users_id_name_index",
	})

	s.blueprint.indexCommand("index", []string{"id", "name"}, schema.IndexConfig{
		Algorithm: "custom_algorithm",
		Name:      "custom_name",
	})
	s.Contains(s.blueprint.commands, &schema.Command{
		Algorithm: "custom_algorithm",
		Columns:   []string{"id", "name"},
		Name:      "index",
		Index:     "custom_name",
	})
}

func (s *BlueprintTestSuite) TestIsCreate() {
	s.False(s.blueprint.isCreate())
	s.blueprint.commands = []*schema.Command{
		{
			Name: constants.CommandCreate,
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
}

func (s *BlueprintTestSuite) TestMediumIncrements() {
	name := "name"
	s.blueprint.MediumIncrements(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("mediumInteger"),
	})
}

func (s *BlueprintTestSuite) TestMediumInteger() {
	name := "name"
	s.blueprint.MediumInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer("mediumInteger"),
	})
}

func (s *BlueprintTestSuite) TestSmallIncrements() {
	name := "name"
	s.blueprint.SmallIncrements(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("smallInteger"),
	})
}

func (s *BlueprintTestSuite) TestSmallInteger() {
	name := "name"
	s.blueprint.SmallInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer("smallInteger"),
	})
}

func (s *BlueprintTestSuite) TestString() {
	column := "name"
	customLength := 100
	length := constants.DefaultStringLength
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

func (s *BlueprintTestSuite) TestTinyIncrements() {
	name := "name"
	s.blueprint.TinyIncrements(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("tinyInteger"),
	})
}

func (s *BlueprintTestSuite) TestTinyInteger() {
	name := "name"
	s.blueprint.TinyInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer("tinyInteger"),
	})
}

func (s *BlueprintTestSuite) TestToSql() {
	for driver, grammar := range s.grammars {
		// Create a table
		s.blueprint.Create()
		s.blueprint.String("name")
		// TODO Add below when implementing the comment method
		//s.blueprint.String("name").Comment("comment")
		//s.blueprint.Comment("comment")

		if driver == database.DriverPostgres {
			s.Len(s.blueprint.ToSql(grammar), 1)
		} else {
			s.Empty(s.blueprint.ToSql(grammar))
		}

		// Update a table
		s.SetupTest()
		s.blueprint.String("avatar")
		if driver == database.DriverPostgres {
			s.Len(s.blueprint.ToSql(grammar), 1)
		} else {
			s.Empty(s.blueprint.ToSql(grammar))
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

func (s *BlueprintTestSuite) TestUnsignedInteger() {
	name := "name"
	s.blueprint.UnsignedInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:     &name,
		ttype:    convert.Pointer("integer"),
		unsigned: convert.Pointer(true),
	})
}

func (s *BlueprintTestSuite) TestUnsignedMediumInteger() {
	name := "name"
	s.blueprint.UnsignedMediumInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:     &name,
		ttype:    convert.Pointer("mediumInteger"),
		unsigned: convert.Pointer(true),
	})
}

func (s *BlueprintTestSuite) TestUnsignedSmallInteger() {
	name := "name"
	s.blueprint.UnsignedSmallInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:     &name,
		ttype:    convert.Pointer("smallInteger"),
		unsigned: convert.Pointer(true),
	})
}

func (s *BlueprintTestSuite) TestUnsignedTinyInteger() {
	name := "name"
	s.blueprint.UnsignedTinyInteger(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:     &name,
		ttype:    convert.Pointer("tinyInteger"),
		unsigned: convert.Pointer(true),
	})
}
