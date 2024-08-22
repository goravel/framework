package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

func TestDefaultDriverCreate(t *testing.T) {
	now := carbon.FromDateTime(2024, 8, 17, 21, 45, 1)
	carbon.SetTestNow(now)
	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "database", "migrations")

	tests := []struct {
		name     string
		argument string
		file     string
		content  string
	}{
		{
			name:     "empty template",
			argument: "fix_users_table",
			file:     "20240817214501_fix_users_table",
			content: `package migrations

type M20240817214501FixUsersTable struct {
}

// Signature The unique signature for the migration.
func (r *M20240817214501FixUsersTable) Signature() string {
	return "20240817214501_fix_users_table"
}

// Connection The database connection that should be used by the migration.
func (r *M20240817214501FixUsersTable) Connection() string {
	return ""
}

// Up Run the migrations.
func (r *M20240817214501FixUsersTable) Up() {

}

// Down Reverse the migrations.
func (r *M20240817214501FixUsersTable) Down() {

}`,
		},
		{
			name:     "create template",
			argument: "create_users_table",
			file:     "20240817214501_create_users_table",
			content: `package migrations

type M20240817214501CreateUsersTable struct {
}

// Signature The unique signature for the migration.
func (r *M20240817214501CreateUsersTable) Signature() string {
	return "20240817214501_create_users_table"
}

// Connection The database connection that should be used by the migration.
func (r *M20240817214501CreateUsersTable) Connection() string {
	return ""
}

// Up Run the migrations.
func (r *M20240817214501CreateUsersTable) Up() {

}

// Down Reverse the migrations.
func (r *M20240817214501CreateUsersTable) Down() {

}`,
		},
		{
			name:     "update template",
			argument: "add_name_to_users_table",
			file:     "20240817214501_add_name_to_users_table",
			content: `package migrations

type M20240817214501AddNameToUsersTable struct {
}

// Signature The unique signature for the migration.
func (r *M20240817214501AddNameToUsersTable) Signature() string {
	return "20240817214501_add_name_to_users_table"
}

// Connection The database connection that should be used by the migration.
func (r *M20240817214501AddNameToUsersTable) Connection() string {
	return ""
}

// Up Run the migrations.
func (r *M20240817214501AddNameToUsersTable) Up() {

}

// Down Reverse the migrations.
func (r *M20240817214501AddNameToUsersTable) Down() {

}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := &DefaultDriver{}

			assert.Nil(t, driver.Create(test.argument))
			assert.Equal(t, test.file, driver.getFileName(test.argument))
			assert.True(t, file.Exists(filepath.Join(path, test.file+".go")))
			assert.True(t, file.Contain(driver.getPath(driver.getFileName(test.argument)), test.content))
		})
	}

	assert.Nil(t, file.Remove("database"))
}
