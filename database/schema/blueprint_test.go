package schema

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/grammars"
	ormmock "github.com/goravel/framework/mocks/database/orm"
	mockschema "github.com/goravel/framework/mocks/database/schema"
	"github.com/goravel/framework/support/convert"
)

type BlueprintTestSuite struct {
	suite.Suite
	blueprint *Blueprint
	grammars  map[ormcontract.Driver]schema.Grammar
}

func TestBlueprintTestSuite(t *testing.T) {
	suite.Run(t, &BlueprintTestSuite{
		grammars: map[ormcontract.Driver]schema.Grammar{
			ormcontract.DriverPostgres: grammars.NewPostgres(),
		},
	})
}

func (s *BlueprintTestSuite) SetupTest() {
	s.blueprint = NewBlueprint("goravel_", "users")
}

func (s *BlueprintTestSuite) TestAddAttributeCommands() {
	var (
		mockGrammar      *mockschema.Grammar
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
			name:  "Should not add command when columns is empty",
			setup: func() {},
		},
		{
			name:    "Should not add command when columns is not empty but GetAttributeCommands does not contain a valid command",
			columns: []*ColumnDefinition{columnDefinition},
			setup: func() {
				mockGrammar.On("GetAttributeCommands").Return([]string{"test"}).Once()
			},
		},
		{
			name:    "Should add comment command when columns is not empty and GetAttributeCommands contains a comment command",
			columns: []*ColumnDefinition{columnDefinition},
			setup: func() {
				mockGrammar.On("GetAttributeCommands").Return([]string{"comment"}).Once()
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
			mockGrammar = &mockschema.Grammar{}
			s.blueprint.columns = test.columns
			test.setup()

			s.blueprint.addAttributeCommands(mockGrammar)
			s.Equal(test.expectCommands, s.blueprint.commands)

			mockGrammar.AssertExpectations(s.T())
		})
	}
}

func (s *BlueprintTestSuite) TestAddImpliedCommands() {
	var (
		mockGrammar *mockschema.Grammar
	)

	tests := []struct {
		name           string
		columns        []*ColumnDefinition
		commands       []*schema.Command
		setup          func()
		expectCommands []*schema.Command
	}{
		{
			name: "Should not add the add command when there are added columns but it is a create operation",
			columns: []*ColumnDefinition{
				{
					name: convert.Pointer("name"),
				},
			},
			commands: []*schema.Command{
				{
					Name: "create",
				},
			},
			setup: func() {
				mockGrammar.On("GetAttributeCommands").Return([]string{}).Once()
			},
			expectCommands: []*schema.Command{
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
			commands: []*schema.Command{
				{
					Name: "create",
				},
			},
			setup: func() {
				mockGrammar.On("GetAttributeCommands").Return([]string{}).Once()
			},
			expectCommands: []*schema.Command{
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
				mockGrammar.On("GetAttributeCommands").Return([]string{"comment"}).Once()
			},
			expectCommands: []*schema.Command{
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
			mockGrammar = &mockschema.Grammar{}
			s.blueprint.columns = test.columns
			s.blueprint.commands = test.commands
			test.setup()

			s.blueprint.addImpliedCommands(mockGrammar)
			s.Equal(test.expectCommands, s.blueprint.commands)

			mockGrammar.AssertExpectations(s.T())
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

	s.blueprint.BigInteger(name, schema.IntegerConfig{
		AutoIncrement: true,
		Unsigned:      true,
	})
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("bigInteger"),
	})
}

func (s *BlueprintTestSuite) TestBuild() {
	for _, grammar := range s.grammars {
		mockQuery := &ormmock.Query{}

		s.blueprint.Create()
		s.blueprint.String("name")

		mockQuery.On("Exec", s.blueprint.ToSql(mockQuery, grammar)[0]).Return(nil, nil).Once()
		s.Nil(s.blueprint.Build(mockQuery, grammar))

		mockQuery.On("Exec", s.blueprint.ToSql(mockQuery, grammar)[0]).Return(nil, errors.New("error")).Once()
		s.EqualError(s.blueprint.Build(mockQuery, grammar), "error")

		mockQuery.AssertExpectations(s.T())
	}
}

func (s *BlueprintTestSuite) TestChar() {
	column := "name"
	customLength := 100
	length := defaultStringLength
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
}

func (s *BlueprintTestSuite) TestDecimal() {
	column := "name"
	ttype := "decimal"

	tests := []struct {
		name          string
		decimalLength *schema.DecimalConfig
		expectColumns []*ColumnDefinition
		expectPlaces  int
		expectTotal   int
	}{
		{
			name:          "decimalLength is nil",
			decimalLength: nil,
			expectColumns: []*ColumnDefinition{
				{
					name:  &column,
					ttype: &ttype,
				},
			},
			expectPlaces: 2,
			expectTotal:  8,
		},
		{
			name:          "decimalLength only with places",
			decimalLength: &schema.DecimalConfig{Places: 4},
			expectColumns: []*ColumnDefinition{
				{
					name:   &column,
					places: convert.Pointer(4),
					total:  convert.Pointer(0),
					ttype:  &ttype,
				},
			},
			expectPlaces: 4,
			expectTotal:  0,
		},
		{
			name:          "decimalLength only with total",
			decimalLength: &schema.DecimalConfig{Total: 10},
			expectColumns: []*ColumnDefinition{
				{
					name:   &column,
					places: convert.Pointer(0),
					total:  convert.Pointer(10),
					ttype:  &ttype,
				},
			},
			expectPlaces: 0,
			expectTotal:  10,
		},
		{
			name:          "decimalLength with total",
			decimalLength: &schema.DecimalConfig{Places: 4, Total: 10},
			expectColumns: []*ColumnDefinition{
				{
					name:   &column,
					places: convert.Pointer(4),
					total:  convert.Pointer(10),
					ttype:  &ttype,
				},
			},
			expectPlaces: 4,
			expectTotal:  10,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.blueprint.columns = []*ColumnDefinition{}

			if test.decimalLength != nil {
				s.blueprint.Decimal(column, *test.decimalLength)
			} else {
				s.blueprint.Decimal(column)
			}
			s.Equal(test.expectColumns, s.blueprint.columns)
		})
	}
}

func (s *BlueprintTestSuite) TestFloat() {
	column := "name"
	customPrecision := 100
	ttype := "float"
	s.blueprint.Float(column)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:      &column,
		precision: convert.Pointer(53),
		ttype:     &ttype,
	})

	s.blueprint.Float(column, customPrecision)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:      &column,
		precision: &customPrecision,
		ttype:     &ttype,
	})
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

func (s *BlueprintTestSuite) TestInteger() {
	name := "name"
	s.blueprint.Integer(name)
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		name:  &name,
		ttype: convert.Pointer("integer"),
	})

	s.blueprint.Integer(name, schema.IntegerConfig{
		AutoIncrement: true,
		Unsigned:      true,
	})
	s.Contains(s.blueprint.GetAddedColumns(), &ColumnDefinition{
		autoIncrement: convert.Pointer(true),
		name:          &name,
		unsigned:      convert.Pointer(true),
		ttype:         convert.Pointer("integer"),
	})
}

func (s *BlueprintTestSuite) TestIsCreate() {
	s.False(s.blueprint.isCreate())
	s.blueprint.commands = []*schema.Command{
		{
			Name: "create",
		},
	}
	s.True(s.blueprint.isCreate())
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
		mockQuery := &ormmock.Query{}
		s.blueprint.Create()
		s.blueprint.String("name").Comment("comment")
		s.blueprint.Comment("comment")

		if driver == ormcontract.DriverPostgres {
			s.Len(s.blueprint.ToSql(mockQuery, grammar), 3)
		} else {
			s.Empty(s.blueprint.ToSql(mockQuery, grammar), 2)
		}
	}
}
