package migration

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/carbon"
)

type DefaultCreatorSuite struct {
	suite.Suite
	defaultCreator *Creator
}

func TestDefaultCreatorSuite(t *testing.T) {
	suite.Run(t, &DefaultCreatorSuite{})
}

func (s *DefaultCreatorSuite) SetupTest() {
	s.defaultCreator = NewCreator()
}

func (s *DefaultCreatorSuite) TestPopulateStub() {
	tests := []struct {
		name      string
		stub      string
		signature string
		table     string
		expected  string
	}{
		{
			name:      "Empty stub",
			stub:      Stubs{}.Empty(),
			signature: "202410131203_create_users_table",
			table:     "users",
			expected: `package migrations

type M202410131203CreateUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M202410131203CreateUsersTable) Signature() string {
	return "202410131203_create_users_table"
}

// Up Run the migrations.
func (r *M202410131203CreateUsersTable) Up() error {
	return nil
}

// Down Reverse the migrations.
func (r *M202410131203CreateUsersTable) Down() error {
	return nil
}
`,
		},
		{
			name:      "Create stub",
			stub:      Stubs{}.Create(),
			signature: "202410131203_create_users_table",
			table:     "users",
			expected: `package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M202410131203CreateUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M202410131203CreateUsersTable) Signature() string {
	return "202410131203_create_users_table"
}

// Up Run the migrations.
func (r *M202410131203CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.ID()
			table.TimestampsTz()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M202410131203CreateUsersTable) Down() error {
 	return facades.Schema().DropIfExists("users")
}
`,
		},
		{
			name:      "Update stub",
			stub:      Stubs{}.Update(),
			signature: "202410131203_create_users_table",
			expected: `package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type M202410131203CreateUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M202410131203CreateUsersTable) Signature() string {
	return "202410131203_create_users_table"
}

// Up Run the migrations.
func (r *M202410131203CreateUsersTable) Up() error {
	return facades.Schema().Table("users", func(table schema.Blueprint) {

	})
}

// Down Reverse the migrations.
func (r *M202410131203CreateUsersTable) Down() error {
	return nil
}
`,
			table: "users",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			actual := s.defaultCreator.PopulateStub(test.stub, test.signature, test.table)
			s.Equal(test.expected, actual)
		})
	}
}

func (s *DefaultCreatorSuite) TestGetFileName() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)

	actual := s.defaultCreator.GetFileName("create_users_table")
	s.Contains(actual, "20240817214501_create_users_table")
}
