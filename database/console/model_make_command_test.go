package console

import (
	"errors"
	"fmt"
	"strings"
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
		{Pattern: "bigint", Type: "uint", NullType: "sql.NullInt64", Imports: []string{"database/sql"}},
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
		{Pattern: "decimal", Type: "float64", NullType: "sql.NullFloat64", Imports: []string{"database/sql"}},
		{Pattern: "int", Type: "int", NullType: "sql.NullInt32", Imports: []string{"database/sql"}},
		{Pattern: "timestamp", Type: "time.Time", NullType: "sql.NullTime", Imports: []string{"database/sql", "time"}},
	}
	mockSchema.EXPECT().GoTypes().Return(goTypes).Once()

	mockContext.EXPECT().Info("Generated 3 fields and 2 embeds from table 'products'").Once()
	mockContext.EXPECT().Success("Model created successfully").Once()

	assert.Nil(t, modelMakeCommand.Handle(mockContext))
	assert.True(t, file.Exists("app/models/product.go"))
	assert.True(t, file.Contain("app/models/product.go", "package models"))
	assert.True(t, file.Contain("app/models/product.go", "type Product struct"))
	assert.True(t, file.Contain("app/models/product.go", "orm.NullableModel"))
	assert.True(t, file.Contain("app/models/product.go", "orm.NullableSoftDeletes"))
	assert.True(t, file.Contain("app/models/product.go", "Name"))
	assert.True(t, file.Contain("app/models/product.go", "Price"))
	assert.True(t, file.Contain("app/models/product.go", "Stock"))
	assert.True(t, file.Contain("app/models/product.go", "TableName"))

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
		{Pattern: "bigint", Type: "uint", NullType: "sql.NullInt64", Imports: []string{"database/sql"}},
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
		{Pattern: "timestamp", Type: "time.Time", NullType: "sql.NullTime", Imports: []string{"database/sql", "time"}},
	}

	mockSchema.EXPECT().GoTypes().Return(goTypes).Once()

	info, err := modelMakeCommand.generateModelInfo(columns, "User", "users")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(info.Fields))
	assert.Equal(t, 2, len(info.Embeds))
	assert.True(t, strings.Contains(info.Fields[0], "Name"))
	assert.True(t, strings.Contains(info.TableNameMethod, "func (r *User) TableName() string"))
	assert.Contains(t, info.Embeds, "orm.Model")
	assert.Contains(t, info.Embeds, "orm.NullableSoftDeletes")
	assert.Contains(t, info.Imports, "github.com/goravel/framework/database/orm")

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
	assert.True(t, strings.Contains(info.Fields[0], "Title"))
	assert.Contains(t, info.Embeds, "orm.NullableTimestamps")
	assert.Contains(t, info.Embeds, "orm.SoftDeletes")

	// Test with no standard columns
	columns = []driver.Column{
		{Name: "name", Type: "varchar", Nullable: false},
		{Name: "price", Type: "decimal", Nullable: false},
	}

	goTypes = []schema.GoType{
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
		{Pattern: "decimal", Type: "float64", NullType: "sql.NullFloat64", Imports: []string{"database/sql"}},
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
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
		{Pattern: "int", Type: "int", NullType: "sql.NullInt32", Imports: []string{"database/sql"}},
	}

	result := getSchemaType("varchar", typeMapping)
	assert.Equal(t, "string", result.Type)
	assert.Equal(t, "sql.NullString", result.NullType)
	assert.Contains(t, result.Imports, "database/sql")

	// Test with regex pattern
	typeMapping = []schema.GoType{
		{Pattern: "varchar.*", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
		{Pattern: "int.*", Type: "int", NullType: "sql.NullInt32", Imports: []string{"database/sql"}},
		{Pattern: "decimal.*", Type: "float64", NullType: "sql.NullFloat64", Imports: []string{"database/sql"}},
	}

	result = getSchemaType("varchar(255)", typeMapping)
	assert.Equal(t, "string", result.Type)
	assert.Equal(t, "sql.NullString", result.NullType)

	result = getSchemaType("integer", typeMapping)
	assert.Equal(t, "int", result.Type)
	assert.Equal(t, "sql.NullInt32", result.NullType)

	result = getSchemaType("decimal(8,2)", typeMapping)
	assert.Equal(t, "float64", result.Type)
	assert.Equal(t, "sql.NullFloat64", result.NullType)

	// Test with enum types
	typeMapping = []schema.GoType{
		{Pattern: "enum", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
		{Pattern: "set", Type: "[]string", NullType: "pq.StringArray", Imports: []string{"github.com/lib/pq"}},
		{Pattern: "json", Type: "map[string]interface{}", NullType: "pq.JsonbArray", Imports: []string{"github.com/lib/pq"}},
	}

	result = getSchemaType("enum('active','inactive','pending')", typeMapping)
	assert.Equal(t, "string", result.Type)
	assert.Equal(t, "sql.NullString", result.NullType)
	assert.Contains(t, result.Imports, "database/sql")

	result = getSchemaType("set('red','green','blue')", typeMapping)
	assert.Equal(t, "[]string", result.Type)
	assert.Equal(t, "pq.StringArray", result.NullType)
	assert.Contains(t, result.Imports, "github.com/lib/pq")

	result = getSchemaType("json", typeMapping)
	assert.Equal(t, "map[string]interface{}", result.Type)
	assert.Equal(t, "pq.JsonbArray", result.NullType)
	assert.Contains(t, result.Imports, "github.com/lib/pq")

	// Test with custom types
	typeMapping = []schema.GoType{
		{Pattern: "uuid", Type: "uuid.UUID", NullType: "uuid.NullUUID", Imports: []string{"github.com/google/uuid"}},
		{Pattern: "geometry", Type: "geo.Point", NullType: "geo.NullPoint", Imports: []string{"github.com/paulmach/go.geo"}},
		{Pattern: "citext", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
	}

	result = getSchemaType("uuid", typeMapping)
	assert.Equal(t, "uuid.UUID", result.Type)
	assert.Equal(t, "uuid.NullUUID", result.NullType)
	assert.Contains(t, result.Imports, "github.com/google/uuid")

	result = getSchemaType("geometry", typeMapping)
	assert.Equal(t, "geo.Point", result.Type)
	assert.Equal(t, "geo.NullPoint", result.NullType)
	assert.Contains(t, result.Imports, "github.com/paulmach/go.geo")

	// Test with no match
	result = getSchemaType("unknown_type", typeMapping)
	assert.Equal(t, "any", result.Type)
	assert.Equal(t, "", result.NullType)
	assert.Empty(t, result.Imports)
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

	result, err := modelMakeCommand.populateStub(stub, "models", "User", modelInfo)
	assert.Nil(t, err)
	assert.Contains(t, result, "package models")
	assert.Contains(t, result, "type User struct")
	assert.Contains(t, result, "ID   uint")
	assert.Contains(t, result, "Name string")
}

func TestGenerateField(t *testing.T) {
	// Test normal field
	column := driver.Column{
		Name:     "name",
		Type:     "varchar",
		Nullable: false,
	}

	typeMapping := []schema.GoType{
		{Pattern: "varchar", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
	}

	field := generateField(column, typeMapping)
	assert.Equal(t, "Name", field.Name)
	assert.Equal(t, "string", field.Type)
	assert.Contains(t, field.Tags, "json:\"name\"")
	assert.Contains(t, field.Tags, "gorm:\"column:name\"")
	assert.Len(t, field.Imports, 1)

	// Test nullable field
	column = driver.Column{
		Name:     "description",
		Type:     "text",
		Nullable: true,
	}

	typeMapping = []schema.GoType{
		{Pattern: "text", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Description", field.Name)
	assert.Equal(t, "sql.NullString", field.Type)
	assert.Contains(t, field.Tags, "json:\"description\"")
	assert.Contains(t, field.Tags, "gorm:\"column:description\"")
	assert.Contains(t, field.Imports, "database/sql")

	// Test auto-increment field
	column = driver.Column{
		Name:          "id",
		Type:          "bigint",
		Nullable:      false,
		Autoincrement: true,
	}

	typeMapping = []schema.GoType{
		{Pattern: "bigint", Type: "uint", NullType: "sql.NullInt64", Imports: []string{"database/sql"}},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Id", field.Name)
	assert.Equal(t, "uint", field.Type)
	assert.Contains(t, field.Tags, "json:\"id\"")
	assert.Contains(t, field.Tags, "gorm:\"column:id;autoIncrement\"")
	assert.Len(t, field.Imports, 1)

	// Test enum field
	column = driver.Column{
		Name:     "status",
		Type:     "enum('active','inactive','pending')",
		Nullable: false,
	}

	typeMapping = []schema.GoType{
		{Pattern: "enum", Type: "string", NullType: "sql.NullString", Imports: []string{"database/sql"}},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Status", field.Name)
	assert.Equal(t, "string", field.Type)
	assert.Contains(t, field.Tags, "json:\"status\"")
	assert.Contains(t, field.Tags, "gorm:\"column:status\"")
	assert.Contains(t, field.Imports, "database/sql")

	// Test custom type field
	column = driver.Column{
		Name:     "location",
		Type:     "geometry",
		Nullable: true,
	}

	typeMapping = []schema.GoType{
		{Pattern: "geometry", Type: "geo.Point", NullType: "geo.NullPoint", Imports: []string{"github.com/paulmach/go.geo"}},
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "Location", field.Name)
	assert.Equal(t, "geo.NullPoint", field.Type)
	assert.Contains(t, field.Tags, "json:\"location\"")
	assert.Contains(t, field.Tags, "gorm:\"column:location\"")
	assert.Contains(t, field.Imports, "github.com/paulmach/go.geo")

	// Test unknown type
	column = driver.Column{
		Name:     "custom_field",
		Type:     "unknown_type",
		Nullable: false,
	}

	field = generateField(column, typeMapping)
	assert.Equal(t, "CustomField", field.Name)
	assert.Equal(t, "any", field.Type)
	assert.Contains(t, field.Tags, "json:\"custom_field\"")
	assert.Contains(t, field.Tags, "gorm:\"column:custom_field\"")
	assert.Empty(t, field.Imports)
}
