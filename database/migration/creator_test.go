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
	data := StubData{
		Package:    "migrations",
		StructName: "M202410131203CreateUsersTable",
		Signature:  "202410131203_create_users_table",
		Table:      "users",
	}

	tests := []struct {
		name        string
		stub        string
		data        StubData
		expected    string
		expectError bool
	}{
		{
			name: "Empty stub",
			stub: Stubs{}.Empty(),
			data: data,
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
			name: "Create stub",
			stub: Stubs{}.Create(),
			data: data,
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
			name: "Create stub with schema fields",
			stub: Stubs{}.Create(),
			data: StubData{
				Package:      "migrations",
				StructName:   "M202410131203CreateUsersTable",
				Signature:    "202410131203_create_users_table",
				Table:        "users",
				SchemaFields: []string{`table.ID()`, `table.String("name")`, `table.TimestampsTz()`},
			},
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
			table.String("name")
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
			name: "Update stub",
			stub: Stubs{}.Update(),
			data: data,
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
		},
		{
			name:        "Invalid template returns error",
			stub:        `{{.InvalidSyntax`,
			data:        data,
			expectError: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual, err := s.defaultCreator.PopulateStub(tc.stub, tc.data)
			if tc.expectError {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tc.expected, actual)
			}
		})
	}
}

func (s *DefaultCreatorSuite) TestGetFileName() {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)

	actual := s.defaultCreator.GetFileName("create_users_table")
	s.Contains(actual, "20240817214501_create_users_table")
}
