package migration

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/database/orm"
	"github.com/goravel/framework/errors"
)

type User struct {
	ID   uint
	Name string
}

type BasicGormModel struct {
	orm.Model
	Name       string
	Email      *string `gorm:"size:255"`
	unexported string
}

type AllFeatures struct {
	orm.Model
	SKU        string `gorm:"column:sku_code;size:100;unique;not null;default:'SKU-001';comment:'Stock Keeping Unit'"`
	Price      uint   `gorm:"type:decimal(10,2);unsigned;index"`
	Status     string `gorm:"type:enum('pending','active','inactive');default:'pending'"`
	Metadata   []byte `gorm:"type:json"`
	OtherData  map[string]any
	Ignored    string `gorm:"-"`
	Relation   User   `gorm:"foreignKey:UserID"`
	Categories []User
}

type CustomTable struct {
	ID uint
}

func (c *CustomTable) TableName() string {
	return "my_custom_products"
}

type CompositeIndexes struct {
	UserID  uint   `gorm:"index:idx_user_zip"`
	ZipCode string `gorm:"index:idx_user_zip;size:10"`
	Email   string `gorm:"uniqueIndex:uix_email_org;size:100"`
	OrgID   uint   `gorm:"uniqueIndex:uix_email_org"`
}

type OtherEmbeds struct {
	orm.Timestamps
	ID   uint
	Name string
}

type JustSoftDelete struct {
	ID   uint
	Name string
	orm.SoftDeletes
}

func TestGenerate(t *testing.T) {
	testCases := []struct {
		name      string
		model     any
		wantTable string
		wantLines []string
		wantErr   error
	}{
		{
			name:      "Basic gorm.Model",
			model:     &BasicGormModel{},
			wantTable: "basic_gorm_models",
			wantLines: []string{
				"table.ID()",
				"table.Text(\"name\")",
				"table.String(\"email\", 255).Nullable()",
				"table.Timestamps()",
				"table.SoftDeletes()",
			},
			wantErr: nil,
		},
		{
			name:      "All Features Model",
			model:     &AllFeatures{},
			wantTable: "all_features",
			wantLines: []string{
				"table.ID()",
				"table.String(\"sku_code\", 100).Default(\"SKU-001\").Comment(\"Stock Keeping Unit\")", // Note: default/comment strings are quoted
				"table.Decimal(\"price\").Unsigned().Total(10).Places(2)",
				"table.Enum(\"status\", []any{\"pending\", \"active\", \"inactive\"}).Default(\"pending\")",
				"table.Json(\"metadata\")",
				"table.Json(\"other_data\")",
				"table.Timestamps()",
				"table.SoftDeletes()",
				"", // Separator line for indexes
				"table.Unique(\"sku_code\")",
				"table.Index(\"price\")",
			},
			wantErr: nil,
		},
		{
			name:      "Custom Table Name",
			model:     &CustomTable{},
			wantTable: "my_custom_products",
			wantLines: []string{
				"table.ID()",
			},
			wantErr: nil,
		},
		{
			name:      "Composite Indexes",
			model:     &CompositeIndexes{},
			wantTable: "composite_indexes",
			wantLines: []string{
				"table.UnsignedInteger(\"user_id\")",
				"table.String(\"zip_code\", 10)",
				"table.String(\"email\", 100)",
				"table.UnsignedInteger(\"org_id\")",
				"", // Separator
				"table.Index(\"user_id\", \"zip_code\")",
				"table.Unique(\"email\", \"org_id\")",
			},
			wantErr: nil,
		},
		{
			name:      "ORM Timestamps Only",
			model:     &OtherEmbeds{},
			wantTable: "other_embeds",
			wantLines: []string{
				"table.ID()",
				"table.Text(\"name\")",
				"table.Timestamps()",
			},
			wantErr: nil,
		},
		{
			name:      "ORM SoftDeletes Only",
			model:     &JustSoftDelete{},
			wantTable: "just_soft_deletes",
			wantLines: []string{
				"table.ID()",
				"table.Text(\"name\")",
				"table.SoftDeletes()",
			},
			wantErr: nil,
		},
		{
			name:      "Error - Nil model",
			model:     nil,
			wantTable: "",
			wantLines: nil,
			wantErr:   errors.SchemaInvalidModel,
		},
		{
			name:      "Error - Non-struct model (int)",
			model:     123,
			wantTable: "",
			wantLines: nil,
			wantErr:   errors.SchemaInvalidModel,
		},
		{
			name:      "Error - Non-struct model (string)",
			model:     "hello",
			wantTable: "",
			wantLines: nil,
			wantErr:   errors.SchemaInvalidModel,
		},
		{
			name:      "Error - Empty struct model",
			model:     struct{}{},
			wantTable: "",
			wantLines: nil,
			wantErr:   errors.SchemaInvalidModel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotTable, gotLines, gotErr := Generate(tc.model)

			assert.ErrorIs(t, gotErr, tc.wantErr)
			assert.Equal(t, tc.wantTable, gotTable)

			sort.Strings(gotLines)
			sort.Strings(tc.wantLines)

			assert.Equal(t, tc.wantLines, gotLines)
		})
	}
}
