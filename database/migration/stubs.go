package migration

type Stubs struct {
}

func (receiver Stubs) Empty() string {
	return `package {{.Package}}

type {{.StructName}} struct{}

// Signature The unique signature for the migration.
func (r *{{.StructName}}) Signature() string {
	return "{{.Signature}}"
}

// Up Run the migrations.
func (r *{{.StructName}}) Up() error {
	return nil
}

// Down Reverse the migrations.
func (r *{{.StructName}}) Down() error {
	return nil
}
`
}

func (receiver Stubs) Create() string {
	return `package {{.Package}}

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type {{.StructName}} struct{}

// Signature The unique signature for the migration.
func (r *{{.StructName}}) Signature() string {
	return "{{.Signature}}"
}

// Up Run the migrations.
func (r *{{.StructName}}) Up() error {
	if !facades.Schema().HasTable("{{.Table}}") {
		return facades.Schema().Create("{{.Table}}", func(table schema.Blueprint) {
			{{- if .SchemaFields}}
			{{range .SchemaFields}}
			{{.}}
			{{- end}}
			{{- else}}
			table.ID()
			table.TimestampsTz()
			{{- end}}
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *{{.StructName}}) Down() error {
 	return facades.Schema().DropIfExists("{{.Table}}")
}
`
}

func (receiver Stubs) Update() string {
	return `package {{.Package}}

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/facades"
)

type {{.StructName}} struct{}

// Signature The unique signature for the migration.
func (r *{{.StructName}}) Signature() string {
	return "{{.Signature}}"
}

// Up Run the migrations.
func (r *{{.StructName}}) Up() error {
	return facades.Schema().Table("{{.Table}}", func(table schema.Blueprint) {
		{{- if .SchemaFields}}
		{{range .SchemaFields}}
		{{.}}
		{{- end}}
		{{- end}}
	})
}

// Down Reverse the migrations.
func (r *{{.StructName}}) Down() error {
	return nil
}
`
}
