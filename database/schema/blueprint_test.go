package schema

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/schema/grammars"
	ormmock "github.com/goravel/framework/mocks/database/orm"
)

type BlueprintTestSuite struct {
	suite.Suite
	blueprint *Blueprint
	grammars  []schema.Grammar
}

func TestBlueprintTestSuite(t *testing.T) {
	suite.Run(t, &BlueprintTestSuite{
		grammars: []schema.Grammar{
			grammars.NewPostgres(),
		},
	})
}

func (s *BlueprintTestSuite) SetupTest() {
	s.blueprint = NewBlueprint("goravel_", "users")
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
	for _, grammar := range s.grammars {
		mockQuery := &ormmock.Query{}
		s.blueprint.Create()
		s.blueprint.String("name")
		s.NotEmpty(s.blueprint.ToSql(mockQuery, grammar))
	}
}
