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

func (s *PostgresSuite) TestCompileChange() {
	var (
		mockBlueprint *mockschema.Blueprint
		mockColumn1   *mockschema.ColumnDefinition
		mockColumn2   *mockschema.ColumnDefinition
	)

	tests := []struct {
		name      string
		setup     func()
		expectSql string
	}{
		{
			name: "no changes",
			setup: func() {
				mockBlueprint.EXPECT().GetChangedColumns().Return([]schemacontract.ColumnDefinition{}).Once()
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
				mockBlueprint.EXPECT().GetChangedColumns().Return([]schemacontract.ColumnDefinition{mockColumn1}).Once()
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
				mockBlueprint.EXPECT().GetChangedColumns().Return([]schemacontract.ColumnDefinition{mockColumn1, mockColumn2}).Once()
			},
			expectSql: "alter table users alter column name set default 'goravel', alter column name drop not null, alter column age set default '1', alter column age set not null",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			mockBlueprint = &mockschema.Blueprint{}
			mockColumn1 = &mockschema.ColumnDefinition{}
			mockColumn2 = &mockschema.ColumnDefinition{}

			test.setup()

			sql := s.grammar.CompileChange(mockBlueprint)

			s.Equal(test.expectSql, sql)

			mockBlueprint.AssertExpectations(s.T())
			mockColumn1.AssertExpectations(s.T())
			mockColumn2.AssertExpectations(s.T())
		})
	}
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

func (s *PostgresSuite) TestCompileIndex() {
	mockBlueprint := &mockschema.Blueprint{}
	mockBlueprint.On("GetTableName").Return("users").Twice()

	sql := s.grammar.CompileIndex(mockBlueprint, &schemacontract.Command{
		Columns: []string{"id", "name"},
		Index:   "id_name",
	})

	s.Equal("create index id_name on users (id, name)", sql)

	sql = s.grammar.CompileIndex(mockBlueprint, &schemacontract.Command{
		Algorithm: "btree",
		Columns:   []string{"id", "name"},
		Index:     "id_name",
	})

	s.Equal("create index id_name on users using btree (id, name)", sql)

	mockBlueprint.AssertExpectations(s.T())
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

func (s *PostgresSuite) TestTypeFloat() {
	mockColumn1 := &mockschema.ColumnDefinition{}
	mockColumn1.On("GetPrecision").Return(100).Once()

	s.Equal("float(100)", s.grammar.TypeFloat(mockColumn1))

	mockColumn2 := &mockschema.ColumnDefinition{}
	mockColumn2.On("GetPrecision").Return(0).Once()

	s.Equal("float", s.grammar.TypeFloat(mockColumn2))
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
