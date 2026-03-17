package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	mocksorm "github.com/goravel/framework/mocks/database/orm"
)

type DBRulesTestSuite struct {
	suite.Suite
	mockOrm   *mocksorm.Orm
	mockQuery *mocksorm.Query
}

func TestDBRulesTestSuite(t *testing.T) {
	suite.Run(t, new(DBRulesTestSuite))
}

func (s *DBRulesTestSuite) SetupTest() {
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.mockQuery = mocksorm.NewQuery(s.T())
	ormFacade = s.mockOrm
}

func (s *DBRulesTestSuite) TearDownTest() {
	ormFacade = nil
}

// --- exists rule tests ---

func (s *DBRulesTestSuite) TestRuleExists_SingleColumn_Found() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(true, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email"},
	}
	s.True(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_SingleColumn_NotFound() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "notfound@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(false, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "notfound@example.com",
		Parameters: []string{"users", "email"},
	}
	s.False(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_DefaultColumn() {
	// When no column specified, defaults to field name
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(true, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users"},
	}
	s.True(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_MultipleColumns_OR() {
	// exists:users,email,username — WHERE email = value OR username = value
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().OrWhere("username", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(true, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "username"},
	}
	s.True(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_MultipleColumns_ThreeFields() {
	// exists:users,email,username,phone
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "value").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().OrWhere("username", "value").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().OrWhere("phone", "value").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(false, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "value",
		Parameters: []string{"users", "email", "username", "phone"},
	}
	s.False(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_ConnectionTable() {
	// exists:mysql.users,email — specify connection
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Connection("mysql").Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Exists().Return(true, nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"mysql.users", "email"},
	}
	s.True(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_OrmNil() {
	ormFacade = nil

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email"},
	}
	s.False(ruleExists(ctx))
}

func (s *DBRulesTestSuite) TestRuleExists_NoParameters() {
	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{},
	}
	// No table specified, should return false
	s.False(ruleExists(ctx))
}

// --- unique rule tests ---

func (s *DBRulesTestSuite) TestRuleUnique_IsUnique() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_NotUnique() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "taken@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(1), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "taken@example.com",
		Parameters: []string{"users", "email"},
	}
	s.False(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_DefaultColumn() {
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_WithExcept() {
	// unique:users,email,id,5 — exclude record where id=5
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().WhereNotIn("id", []any{"5"}).Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "id", "5"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_WithCustomIdColumnAndExcept() {
	// unique:users,email,user_id,5 — exclude record where user_id=5
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().WhereNotIn("user_id", []any{"5"}).Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "user_id", "5"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_WithMultipleExcepts() {
	// unique:users,email,id,1,2,3 — exclude records where id IN (1, 2, 3)
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().WhereNotIn("id", []any{"1", "2", "3"}).Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "id", "1", "2", "3"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_WithDefaultIdColumn() {
	// unique:users,email,,5 — idColumn defaults to "id"
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().WhereNotIn("id", []any{"5"}).Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email", "", "5"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_ConnectionTable() {
	// unique:pgsql.users,email
	s.mockOrm.EXPECT().WithContext(mock.Anything).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Connection("pgsql").Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Table("users").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Where("email", "test@example.com").Return(s.mockQuery).Once()
	s.mockQuery.EXPECT().Count().Return(int64(0), nil).Once()

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"pgsql.users", "email"},
	}
	s.True(ruleUnique(ctx))
}

func (s *DBRulesTestSuite) TestRuleUnique_OrmNil() {
	ormFacade = nil

	ctx := &RuleContext{
		Ctx:        context.Background(),
		Attribute:  "email",
		Value:      "test@example.com",
		Parameters: []string{"users", "email"},
	}
	s.False(ruleUnique(ctx))
}

// --- parseExistsParams tests ---

func (s *DBRulesTestSuite) TestParseExistsParams() {
	tests := []struct {
		name          string
		attribute     string
		parameters    []string
		expectedTable string
		expectedCols  []string
		expectedConn  string
	}{
		{
			name:          "no parameters",
			attribute:     "email",
			parameters:    []string{},
			expectedTable: "",
			expectedCols:  []string{"email"},
			expectedConn:  "",
		},
		{
			name:          "table only",
			attribute:     "email",
			parameters:    []string{"users"},
			expectedTable: "users",
			expectedCols:  []string{"email"},
			expectedConn:  "",
		},
		{
			name:          "table and column",
			attribute:     "email",
			parameters:    []string{"users", "user_email"},
			expectedTable: "users",
			expectedCols:  []string{"user_email"},
			expectedConn:  "",
		},
		{
			name:          "table and multiple columns",
			attribute:     "email",
			parameters:    []string{"users", "email", "username", "phone"},
			expectedTable: "users",
			expectedCols:  []string{"email", "username", "phone"},
			expectedConn:  "",
		},
		{
			name:          "connection.table",
			attribute:     "email",
			parameters:    []string{"mysql.users", "email"},
			expectedTable: "users",
			expectedCols:  []string{"email"},
			expectedConn:  "mysql",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := &RuleContext{
				Attribute:  tt.attribute,
				Parameters: tt.parameters,
			}
			table, cols, conn := parseExistsParams(ctx)
			s.Equal(tt.expectedTable, table)
			s.Equal(tt.expectedCols, cols)
			s.Equal(tt.expectedConn, conn)
		})
	}
}

// --- parseUniqueParams tests ---

func (s *DBRulesTestSuite) TestParseUniqueParams() {
	tests := []struct {
		name          string
		attribute     string
		parameters    []string
		expectedTable string
		expectedCol   string
		expectedConn  string
	}{
		{
			name:          "no parameters",
			attribute:     "email",
			parameters:    []string{},
			expectedTable: "",
			expectedCol:   "email",
			expectedConn:  "",
		},
		{
			name:          "table only",
			attribute:     "email",
			parameters:    []string{"users"},
			expectedTable: "users",
			expectedCol:   "email",
			expectedConn:  "",
		},
		{
			name:          "table and column",
			attribute:     "email",
			parameters:    []string{"users", "user_email"},
			expectedTable: "users",
			expectedCol:   "user_email",
			expectedConn:  "",
		},
		{
			name:          "connection.table",
			attribute:     "email",
			parameters:    []string{"pgsql.users", "email"},
			expectedTable: "users",
			expectedCol:   "email",
			expectedConn:  "pgsql",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := &RuleContext{
				Attribute:  tt.attribute,
				Parameters: tt.parameters,
			}
			table, col, conn := parseUniqueParams(ctx)
			s.Equal(tt.expectedTable, table)
			s.Equal(tt.expectedCol, col)
			s.Equal(tt.expectedConn, conn)
		})
	}
}
