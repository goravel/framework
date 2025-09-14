package console

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goravel/framework/contracts/database/driver"
	"github.com/goravel/framework/contracts/database/schema"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksschema "github.com/goravel/framework/mocks/database/schema"
	"github.com/goravel/framework/support/file"
)

func TestModelMakeCommand(t *testing.T) {
	mockSchema := mocksschema.NewSchema(t)
	mockArtisan := mocksconsole.NewArtisan(t)

	modelMakeCommand := NewModelMakeCommand(mockArtisan, mockSchema)
	mockContext := mocksconsole.NewContext(t)

	// Test: Empty model name
	mockContext.EXPECT().Argument(0).Return("").Once()
	mockContext.EXPECT().Ask("Enter the model name", mock.Anything).Return("", errors.New("the model name cannot be empty")).Once()
	mockContext.EXPECT().Error("the model name cannot be empty").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.False(t, file.Exists("app/models/user.go"))

	// Test: Create model successfully
	mockContext.EXPECT().Argument(0).Return("User").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Option("table").Return("").Once()
	mockContext.EXPECT().Success("Model created successfully").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/models/user.go"))

	// Verify the entire content of the basic model file
	userModel, err := file.GetContent("app/models/user.go")
	assert.Nil(t, err)
	expectedUserContent := `package models

type User struct {
}
`
	assert.Equal(t, expectedUserContent, userModel)

	// Test: Model already exists
	mockContext.EXPECT().Argument(0).Return("User").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Error("the model already exists. Use the --force or -f flag to overwrite").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))

	// Test: Create model in subdirectory
	mockContext.EXPECT().Argument(0).Return("User/Phone").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Option("table").Return("").Once()
	mockContext.EXPECT().Success("Model created successfully").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/models/User/phone.go"))
	assert.True(t, file.Contain("app/models/User/phone.go", "package User"))
	assert.True(t, file.Contain("app/models/User/phone.go", "type Phone struct"))

	// Verify the entire content of the model in subdirectory
	phoneModel, err := file.GetContent("app/models/User/phone.go")
	assert.Nil(t, err)
	expectedPhoneContent := `package User

type Phone struct {
}
`
	assert.Equal(t, expectedPhoneContent, phoneModel)

	// Test: Create model from table schema
	mockContext.EXPECT().Argument(0).Return("Product").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Option("table").Return("products").Once()

	// Table exists
	mockSchema.EXPECT().HasTable("products").Return(true).Once()

	// Mock column data
	columns := []driver.Column{
		{Name: "id", Type: "bigint", Nullable: false, Autoincrement: true},
		{Name: "name", Type: "varchar", Nullable: false},
		{Name: "price", Type: "decimal", Nullable: false},
		{Name: "stock", Type: "int", Nullable: true},
		{Name: "created_at", Type: "timestamp", Nullable: true},
		{Name: "updated_at", Type: "timestamp", Nullable: true},
		{Name: "deleted_at", Type: "timestamp", Nullable: true},
	}
	mockSchema.EXPECT().GetColumns("products").Return(columns, nil).Once()

	// Mock GoTypes
	goTypes := []schema.GoType{
		{Pattern: "bigint", Type: "uint", NullType: "sql.NullInt64", NullImport: "database/sql"},
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", NullImport: "database/sql"},
		{Pattern: "decimal", Type: "float64", NullType: "sql.NullFloat64", NullImport: "database/sql"},
		{Pattern: "int", Type: "int", NullType: "sql.NullInt32", NullImport: "database/sql"},
		{Pattern: "timestamp", Type: "time.Time", NullType: "sql.NullTime", NullImport: "database/sql", Import: "time"},
	}
	mockSchema.EXPECT().GoTypes().Return(goTypes).Once()

	mockContext.EXPECT().Success("Model created successfully").Once()

	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	model, err := file.GetContent("app/models/product.go")
	assert.Nil(t, err)
	expectedContent := `package models

import (
	"database/sql"
	"github.com/goravel/framework/database/orm"
)

type Product struct {
	orm.Model
	orm.SoftDeletes
	Name  string        ` + "`json:\"name\" db:\"name\"`" + `
	Price float64       ` + "`json:\"price\" db:\"price\"`" + `
	Stock sql.NullInt32 ` + "`json:\"stock\" db:\"stock\"`" + `
}

func (r *Product) TableName() string {
	return "products"
}
`
	assert.Equal(t, expectedContent, model)

	// Test: Table doesn't exist
	mockContext.EXPECT().Argument(0).Return("Invalid").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Option("table").Return("nonexistent").Once()
	mockSchema.EXPECT().HasTable("nonexistent").Return(false).Once()
	mockContext.EXPECT().Error("table nonexistent not found").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))

	//Test: Error fetching columns
	mockContext.EXPECT().Argument(0).Return("Error").Once()
	mockContext.EXPECT().OptionBool("force").Return(false).Once()
	mockContext.EXPECT().Option("table").Return("error_table").Once()
	mockSchema.EXPECT().HasTable("error_table").Return(true).Once()
	mockSchema.EXPECT().GetColumns("error_table").Return(nil, errors.New("database connection error")).Once()
	mockContext.EXPECT().Error("database connection error").Once()
	assert.Nil(t, modelMakeCommand.Handle(mockContext))

	assert.Nil(t, file.Remove("app"))
}

func TestGenerateModelInfo(t *testing.T) {
	mockSchema := mocksschema.NewSchema(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	modelMakeCommand := NewModelMakeCommand(mockArtisan, mockSchema)

	// Test with standard columns (id, timestamps, soft deletes)
	columns := []driver.Column{
		{Name: "id", Type: "bigint", Nullable: false, Autoincrement: true},
		{Name: "name", Type: "varchar", Nullable: false},
		{Name: "created_at", Type: "timestamp", Nullable: false},
		{Name: "updated_at", Type: "timestamp", Nullable: false},
		{Name: "deleted_at", Type: "timestamp", Nullable: true},
	}

	goTypes := []schema.GoType{
		{Pattern: "bigint", Type: "uint", NullType: "sql.NullInt64", NullImport: "database/sql"},
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", NullImport: "database/sql"},
		{Pattern: "timestamp", Type: "time.Time", NullType: "sql.NullTime", NullImport: "database/sql", Import: "time"},
	}

	mockSchema.EXPECT().GoTypes().Return(goTypes).Once()

	info, err := modelMakeCommand.generateModelInfo(columns, "User", "users")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(info.Fields))
	assert.Equal(t, 2, len(info.Embeds))
	assert.Contains(t, info.Fields[0], "Name")
	assert.Equal(t, "func (r *User) TableName() string {\n\treturn \"users\"\n}", info.TableNameMethod)
	assert.Equal(t, []string{
		"orm.Model",
		"orm.SoftDeletes",
	}, info.Embeds)
	assert.Equal(t, map[string]struct{}{
		"database/sql": {},
		"github.com/goravel/framework/database/orm": {},
	}, info.Imports)

	// Test with only nullable timestamps (no id, with soft deletes)
	columns = []driver.Column{
		{Name: "title", Type: "varchar", Nullable: false},
		{Name: "created_at", Type: "timestamp", Nullable: true},
		{Name: "updated_at", Type: "timestamp", Nullable: true},
		{Name: "deleted_at", Type: "timestamp", Nullable: false},
	}

	mockSchema.EXPECT().GoTypes().Return(goTypes).Once()

	info, err = modelMakeCommand.generateModelInfo(columns, "Post", "posts")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(info.Fields))
	assert.Equal(t, 2, len(info.Embeds))
	assert.Contains(t, info.Fields[0], "Title")
	assert.Equal(t, []string{
		"orm.Timestamps",
		"orm.SoftDeletes",
	}, info.Embeds)

	// Test with no standard columns
	columns = []driver.Column{
		{Name: "name", Type: "varchar", Nullable: false},
		{Name: "price", Type: "decimal", Nullable: false},
	}

	goTypes = []schema.GoType{
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", NullImport: "database/sql"},
		{Pattern: "decimal", Type: "float64", NullType: "sql.NullFloat64", NullImport: "database/sql"},
	}

	mockSchema.EXPECT().GoTypes().Return(goTypes).Once()

	info, err = modelMakeCommand.generateModelInfo(columns, "Product", "products")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(info.Fields))
	assert.Equal(t, 0, len(info.Embeds))
}

func TestBuildTableNameMethod(t *testing.T) {
	mockSchema := mocksschema.NewSchema(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	modelMakeCommand := NewModelMakeCommand(mockArtisan, mockSchema)

	// Test with standard naming
	result := modelMakeCommand.buildTableNameMethod("User", "users")
	expected := "func (r *User) TableName() string {\n\treturn \"users\"\n}"
	assert.Equal(t, expected, result)

	// Test with custom table name
	result = modelMakeCommand.buildTableNameMethod("Product", "store_items")
	expected = "func (r *Product) TableName() string {\n\treturn \"store_items\"\n}"
	assert.Equal(t, expected, result)
}

func TestBuildField(t *testing.T) {
	mockSchema := mocksschema.NewSchema(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	modelMakeCommand := NewModelMakeCommand(mockArtisan, mockSchema)

	// Test simple field
	result := modelMakeCommand.buildField("Name", "string", "`json:\"name\"`")
	expected := fmt.Sprintf("%-15s %-10s %s", "Name", "string", "`json:\"name\"`")
	assert.Equal(t, expected, result)

	// Test with complex tags
	result = modelMakeCommand.buildField("CreatedAt", "time.Time", "`json:\"created_at\" gorm:\"column:created_at\"`")
	expected = fmt.Sprintf("%-15s %-10s %s", "CreatedAt", "time.Time", "`json:\"created_at\" gorm:\"column:created_at\"`")
	assert.Equal(t, expected, result)
}

func TestGetSchemaType(t *testing.T) {
	// Test with exact match
	typeMapping := []schema.GoType{
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", NullImport: "database/sql"},
		{Pattern: "int", Type: "int", NullType: "sql.NullInt32", NullImport: "database/sql"},
	}

	result := getSchemaType("varchar", typeMapping)
	assert.Equal(t, "string", result.Type)
	assert.Equal(t, "sql.NullString", result.NullType)
	assert.Equal(t, "database/sql", result.NullImport)

	// Test with regex pattern
	typeMapping = []schema.GoType{
		{Pattern: "varchar.*", Type: "string", NullType: "sql.NullString", NullImport: "database/sql"},
		{Pattern: "int.*", Type: "int", NullType: "sql.NullInt32", NullImport: "database/sql"},
		{Pattern: "decimal.*", Type: "float64", NullType: "sql.NullFloat64", NullImport: "database/sql"},
	}

	result = getSchemaType("varchar(255)", typeMapping)
	assert.Equal(t, "string", result.Type)
	assert.Equal(t, "sql.NullString", result.NullType)
	assert.Equal(t, "database/sql", result.NullImport)

	result = getSchemaType("integer", typeMapping)
	assert.Equal(t, "int", result.Type)
	assert.Equal(t, "sql.NullInt32", result.NullType)
	assert.Equal(t, "database/sql", result.NullImport)

	result = getSchemaType("decimal(8,2)", typeMapping)
	assert.Equal(t, "float64", result.Type)
	assert.Equal(t, "sql.NullFloat64", result.NullType)
	assert.Equal(t, "database/sql", result.NullImport)

	// Test with Import field
	typeMapping = []schema.GoType{
		{Pattern: "timestamp", Type: "time.Time", NullType: "sql.NullTime", Import: "time", NullImport: "database/sql"},
		{Pattern: "uuid", Type: "uuid.UUID", NullType: "uuid.NullUUID", Import: "github.com/google/uuid", NullImport: "github.com/google/uuid"},
	}

	result = getSchemaType("timestamp", typeMapping)
	assert.Equal(t, "time.Time", result.Type)
	assert.Equal(t, "sql.NullTime", result.NullType)
	assert.Equal(t, "time", result.Import)
	assert.Equal(t, "database/sql", result.NullImport)

	result = getSchemaType("uuid", typeMapping)
	assert.Equal(t, "uuid.UUID", result.Type)
	assert.Equal(t, "uuid.NullUUID", result.NullType)
	assert.Equal(t, "github.com/google/uuid", result.Import)
	assert.Equal(t, "github.com/google/uuid", result.NullImport)
}

func TestPopulateStub(t *testing.T) {
	mockSchema := mocksschema.NewSchema(t)
	mockArtisan := mocksconsole.NewArtisan(t)
	modelMakeCommand := NewModelMakeCommand(mockArtisan, mockSchema)

	stub := "package {{.PackageName}}\ntype {{.StructName}} struct {\n{{range .Fields}}\t{{.}}\n{{end}}\n}"

	modelInfo := modelDefinition{
		Fields:          []string{"ID uint", "Name string"},
		Imports:         map[string]struct{}{},
		TableNameMethod: "",
	}

	expectedContent := `package models

type User struct {
	ID   uint
	Name string
}
`

	result, err := modelMakeCommand.populateStub(stub, "models", "User", modelInfo)
	assert.Nil(t, err)
	assert.Equal(t, expectedContent, result)
}

func TestGenerateField(t *testing.T) {
	// Test normal field
	column := driver.Column{
		Name:     "name",
		Type:     "varchar",
		Nullable: false,
	}

	typeMapping := []schema.GoType{
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", NullImport: "database/sql"},
	}

	field := generateField(column, typeMapping)
	assert.Equal(t, "Name", field.Name)
	assert.Equal(t, "string", field.Type)
	assert.Equal(t, "`json:\"name\" db:\"name\"`", field.Tags)
	assert.Equal(t, []string{"database/sql"}, field.Imports)

	// Test nullable field
	column = driver.Column{
		Name:     "description",
		Type:     "text",
		Nullable: true,
	}

	typeMapping = []schema.GoType{
		{Pattern: "text", Type: "string", NullType: "sql.NullString", NullImport: "database/sql"},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Description", field.Name)
	assert.Equal(t, "sql.NullString", field.Type)
	assert.Equal(t, "`json:\"description\" db:\"description\"`", field.Tags)
	assert.Equal(t, []string{"database/sql"}, field.Imports)

	// Test auto-increment field
	column = driver.Column{
		Name:          "id",
		Type:          "bigint",
		Nullable:      false,
		Autoincrement: true,
	}

	typeMapping = []schema.GoType{
		{Pattern: "bigint", Type: "uint", NullType: "sql.NullInt64", NullImport: "database/sql"},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Id", field.Name)
	assert.Equal(t, "uint", field.Type)
	assert.Equal(t, "`json:\"id\" db:\"id\" gorm:\"primaryKey\"`", field.Tags)
	assert.Equal(t, []string{"database/sql"}, field.Imports)

	// Test enum field
	column = driver.Column{
		Name:     "status",
		Type:     "enum('active','inactive','pending')",
		Nullable: false,
	}

	typeMapping = []schema.GoType{
		{Pattern: "enum", Type: "string", NullType: "sql.NullString", NullImport: "database/sql"},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Status", field.Name)
	assert.Equal(t, "string", field.Type)
	assert.Equal(t, "`json:\"status\" db:\"status\"`", field.Tags)
	assert.Equal(t, []string{"database/sql"}, field.Imports)

	// Test custom type field
	column = driver.Column{
		Name:     "location",
		Type:     "geometry",
		Nullable: true,
	}

	typeMapping = []schema.GoType{
		{Pattern: "geometry", Type: "geo.Point", NullType: "geo.NullPoint", Import: "github.com/paulmach/go.geo", NullImport: "github.com/paulmach/go.geo"},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Location", field.Name)
	assert.Equal(t, "geo.NullPoint", field.Type)
	assert.Equal(t, "`json:\"location\" db:\"location\"`", field.Tags)
	assert.Contains(t, field.Imports, "github.com/paulmach/go.geo")

	// Test field with Import only (not nullable)
	column = driver.Column{
		Name:     "timestamp",
		Type:     "timestamp",
		Nullable: false,
	}

	typeMapping = []schema.GoType{
		{Pattern: "timestamp", Type: "time.Time", NullType: "sql.NullTime", Import: "time", NullImport: "database/sql"},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Timestamp", field.Name)
	assert.Equal(t, "time.Time", field.Type)
	assert.Equal(t, "`json:\"timestamp\" db:\"timestamp\"`", field.Tags)
	assert.Equal(t, []string{"time", "database/sql"}, field.Imports)

	// Test unknown type
	column = driver.Column{
		Name:     "custom_field",
		Type:     "unknown_type",
		Nullable: false,
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "CustomField", field.Name)
	assert.Equal(t, "any", field.Type)
	assert.Equal(t, "`json:\"custom_field\" db:\"custom_field\"`", field.Tags)
	assert.Empty(t, field.Imports)
}
