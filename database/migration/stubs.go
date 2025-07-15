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
			table.Timestamps()
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

type MysqlStubs struct {
}

// CreateUp Create up migration content.
func (receiver MysqlStubs) CreateUp() string {
	return `CREATE TABLE DummyTable (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_DummyTable_created_at (created_at),
  KEY idx_DummyTable_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = DummyDatabaseCharset;
`
}

// CreateDown Create down migration content.
func (receiver MysqlStubs) CreateDown() string {
	return `DROP TABLE IF EXISTS DummyTable;
`
}

// UpdateUp Update up migration content.
func (receiver MysqlStubs) UpdateUp() string {
	return `ALTER TABLE DummyTable ADD column varchar(255) COMMENT '';
`
}

// UpdateDown Update down migration content.
func (receiver MysqlStubs) UpdateDown() string {
	return `ALTER TABLE DummyTable DROP COLUMN column;
`
}

type PostgresStubs struct {
}

// CreateUp Create up migration content.
func (receiver PostgresStubs) CreateUp() string {
	return `CREATE TABLE DummyTable (
  id SERIAL PRIMARY KEY NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
}

// CreateDown Create down migration content.
func (receiver PostgresStubs) CreateDown() string {
	return `DROP TABLE IF EXISTS DummyTable;
`
}

// UpdateUp Update up migration content.
func (receiver PostgresStubs) UpdateUp() string {
	return `ALTER TABLE DummyTable ADD column varchar(255) NOT NULL;
`
}

// UpdateDown Update down migration content.
func (receiver PostgresStubs) UpdateDown() string {
	return `ALTER TABLE DummyTable DROP COLUMN column;
`
}

type SqliteStubs struct {
}

// CreateUp Create up migration content.
func (receiver SqliteStubs) CreateUp() string {
	return `CREATE TABLE DummyTable (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
}

// CreateDown Create down migration content.
func (receiver SqliteStubs) CreateDown() string {
	return `DROP TABLE IF EXISTS DummyTable;
`
}

// UpdateUp Update up migration content.
func (receiver SqliteStubs) UpdateUp() string {
	return `ALTER TABLE DummyTable ADD column text;
`
}

// UpdateDown Update down migration content.
func (receiver SqliteStubs) UpdateDown() string {
	return `ALTER TABLE DummyTable DROP COLUMN column;
`
}

type SqlserverStubs struct {
}

// CreateUp Create up migration content.
func (receiver SqlserverStubs) CreateUp() string {
	return `CREATE TABLE DummyTable (
  id bigint NOT NULL IDENTITY(1,1),
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
}

// CreateDown Create down migration content.
func (receiver SqlserverStubs) CreateDown() string {
	return `DROP TABLE IF EXISTS DummyTable;
`
}

// UpdateUp Update up migration content.
func (receiver SqlserverStubs) UpdateUp() string {
	return `ALTER TABLE DummyTable ADD column varchar(255);
`
}

// UpdateDown Update down migration content.
func (receiver SqlserverStubs) UpdateDown() string {
	return `ALTER TABLE DummyTable DROP COLUMN column;
`
}
