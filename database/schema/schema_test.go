package schema

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractsschema "github.com/goravel/framework/contracts/database/schema"
)

type User struct {
	ID int64 `gorm:"primaryKey"`
}

type Address struct {
	ID int64 `gorm:"primaryKey"`
}

type Relation struct {
	UserID    int64 `gorm:"primaryKey"`
	AddressID int64 `gorm:"primaryKey"`
}

func defaultModels() []contractsschema.Model {
	return []contractsschema.Model{
		{Name: "User", Type: &User{}},
	}
}

type SchemaTestSuite struct {
	suite.Suite
}

func TestSchemaTestSuite(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}

func (r *SchemaTestSuite) TestExtendGoTypes() {
	defaultLen := len(defaultGoTypes())
	tests := []struct {
		name      string
		overrides []contractsschema.GoType
		assert    func(schema *Schema)
	}{
		{
			name:      "empty_overrides",
			overrides: nil,
			assert: func(schema *Schema) {
				r.Equal(defaultLen, len(schema.goTypes), "goTypes length should remain unchanged with nil overrides")
			},
		},
		{
			name: "add_single_new_type",
			overrides: []contractsschema.GoType{
				{Pattern: "custom_type", Type: "CustomType", NullType: "*CustomType", Import: "github.com/custom/pkg"},
			},
			assert: func(schema *Schema) {
				r.Equal(defaultLen+1, len(schema.goTypes), "goTypes length should increase by 1")

				customType, found := findGoTypeByPattern("custom_type", schema.goTypes)
				r.True(found, "custom_type should be found in goTypes")
				r.Equal("CustomType", customType.Type, "Type should match")
				r.Equal("*CustomType", customType.NullType, "NullType should match")
				r.Equal("github.com/custom/pkg", customType.Import, "Import should match")
			},
		},
		{
			name: "add_multiple_new_types",
			overrides: []contractsschema.GoType{
				{Pattern: "type1", Type: "Type1", NullType: "*Type1", Import: "pkg1"},
				{Pattern: "type2", Type: "Type2", NullType: "*Type2", Import: "pkg2"},
			},
			assert: func(schema *Schema) {
				r.Equal(defaultLen+2, len(schema.goTypes), "goTypes length should increase by 2")

				type1, found := findGoTypeByPattern("type1", schema.goTypes)
				r.True(found, "type1 should be found in goTypes")
				r.Equal("Type1", type1.Type, "Type should match for type1")

				type2, found := findGoTypeByPattern("type2", schema.goTypes)
				r.True(found, "type2 should be found in goTypes")
				r.Equal("Type2", type2.Type, "Type should match for type2")
			},
		},
		{
			name: "override_existing_type_fully",
			overrides: []contractsschema.GoType{
				{Pattern: "(?i)^uuid$", Type: "custom.UUID", NullType: "*custom.UUID", Import: "custom/uuid", NullImport: "custom/uuid/null"},
			},
			assert: func(schema *Schema) {
				r.Equal(defaultLen, len(schema.goTypes), "goTypes length should remain unchanged when overriding")

				// Find the overridden type
				uuidType, found := findGoTypeByPattern("(?i)^uuid$", schema.goTypes)
				r.True(found, "uuid type should be found in goTypes")
				r.Equal("custom.UUID", uuidType.Type, "Type should be overridden")
				r.Equal("*custom.UUID", uuidType.NullType, "NullType should be overridden")
				r.Equal("custom/uuid", uuidType.Import, "Import should be overridden")
				r.Equal("custom/uuid/null", uuidType.NullImport, "NullImport should be overridden")
			},
		},
		{
			name: "override_existing_type_partially",
			overrides: []contractsschema.GoType{
				{Pattern: "(?i)^timestamp$", Type: "time.Time", NullType: "", Import: "time"},
			},
			assert: func(schema *Schema) {
				timestampType, found := findGoTypeByPattern("(?i)^timestamp$", schema.goTypes)
				r.True(found, "timestamp type should be found in goTypes")
				r.Equal("time.Time", timestampType.Type, "Type should be overridden")
				r.Equal("time", timestampType.Import, "Import should be overridden")

				originalTimestamp, _ := findGoTypeByPattern("(?i)^timestamp$", defaultGoTypes())
				r.Equal(originalTimestamp.NullType, timestampType.NullType, "NullType should remain unchanged")
			},
		},
		{
			name: "mix_new_and_override",
			overrides: []contractsschema.GoType{
				{Pattern: "new_pattern", Type: "NewType", NullType: "*NewType"},
				{Pattern: "(?i)^json$", Type: "CustomJSON", NullType: "*CustomJSON"},
			},
			assert: func(schema *Schema) {
				r.Equal(defaultLen+1, len(schema.goTypes), "goTypes length should increase by 1")

				// Check new type
				newType, found := findGoTypeByPattern("new_pattern", schema.goTypes)
				r.True(found, "new_pattern should be found in goTypes")
				r.Equal("NewType", newType.Type, "Type should match for new pattern")

				// Check overridden type
				jsonType, found := findGoTypeByPattern("(?i)^json$", schema.goTypes)
				r.True(found, "json type should be found in goTypes")
				r.Equal("CustomJSON", jsonType.Type, "Type should be overridden for json")
				r.Equal("*CustomJSON", jsonType.NullType, "NullType should be overridden for json")
			},
		},
		{
			name: "multiple_overrides_for_same_pattern",
			overrides: []contractsschema.GoType{
				{Pattern: "(?i)^uuid$", Type: "FirstUUID", NullType: "*FirstUUID"},
				{Pattern: "(?i)^uuid$", Type: "LastUUID", NullType: "*LastUUID"},
			},
			assert: func(schema *Schema) {
				uuidType, found := findGoTypeByPattern("(?i)^uuid$", schema.goTypes)
				r.True(found, "uuid type should be found in goTypes")
				r.Equal("LastUUID", uuidType.Type, "Type should be set to the last override value")
				r.Equal("*LastUUID", uuidType.NullType, "NullType should be set to the last override value")
			},
		},
	}

	for _, test := range tests {
		r.Run(test.name, func() {
			schema := getSchema()
			schema.extendGoTypes(test.overrides)
			test.assert(schema)
		})
	}
}

func (r *SchemaTestSuite) TestExtendModels() {
	defaultLen := len(defaultModels())
	tests := []struct {
		name      string
		overrides []contractsschema.Model
		assert    func(schema *Schema)
	}{
		{
			name:      "empty_overrides",
			overrides: nil,
			assert: func(schema *Schema) {
				r.Equal(defaultLen, len(schema.models), "models length should remain unchanged with nil overrides")
			},
		},
		{
			name: "add_single_new_model",
			overrides: []contractsschema.Model{
				{Name: "Address", Type: &Address{}},
			},
			assert: func(schema *Schema) {
				r.Equal(defaultLen+1, len(schema.models), "models length should increase by 1")

				addressModel := schema.GetModel("Address")
				r.NotNil(addressModel, "Address model should not be nil")
				r.Equal("Address", addressModel.Name, "Model name should match")
			},
		},
		{
			name: "add_multiple_new_models",
			overrides: []contractsschema.Model{
				{Name: "Address", Type: &Address{}},
				{Name: "Relation", Type: &Relation{}},
			},
			assert: func(schema *Schema) {
				r.Equal(defaultLen+2, len(schema.models), "models length should increase by 2")

				addressModel := schema.GetModel("Address")
				r.NotNil(addressModel, "Address model should not be nil")
				r.Equal("Address", addressModel.Name, "Model name should match")

				relationModel := schema.GetModel("Relation")
				r.NotNil(relationModel, "Relation model should not be nil")
				r.Equal("Relation", relationModel.Name, "Model name should match")
			},
		},
		{
			name: "override_existing_model",
			overrides: []contractsschema.Model{
				{Name: "User", Type: &User{}},
			},
			assert: func(schema *Schema) {
				r.Equal(defaultLen, len(schema.models), "models length should remain unchanged when overriding")

				userModel := schema.GetModel("User")
				r.NotNil(userModel, "User model should not be nil")
				r.Equal("User", userModel.Name, "Model name should match")
			},
		},
	}

	for _, test := range tests {
		r.Run(test.name, func() {
			schema := getSchema()
			schema.extendModels(test.overrides)
			test.assert(schema)
		})
	}
}

func getSchema() *Schema {
	return &Schema{
		goTypes: defaultGoTypes(),
		models:  defaultModels(),
	}
}

func findGoTypeByPattern(pattern string, types []contractsschema.GoType) (contractsschema.GoType, bool) {
	for _, t := range types {
		if t.Pattern == pattern {
			return t, true
		}
	}
	return contractsschema.GoType{}, false
}
