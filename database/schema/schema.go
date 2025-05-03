package schema

import (
	"slices"
	"strings"

	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/log"
	"github.com/goravel/framework/errors"
)

var _ contractsschema.Schema = (*Schema)(nil)

type Schema struct {
	config     config.Config
	driver     driver.Driver
	grammar    driver.Grammar
	log        log.Log
	migrations []contractsschema.Migration
	orm        contractsorm.Orm
	prefix     string
	processor  driver.Processor
	schema     string
	goTypes    []contractsschema.GoType
}

func NewSchema(config config.Config, log log.Log, orm contractsorm.Orm, driver driver.Driver, migrations []contractsschema.Migration) (*Schema, error) {
	writers := driver.Pool().Writers
	if len(writers) == 0 {
		return nil, errors.DatabaseConfigNotFound
	}

	prefix := writers[0].Prefix
	schema := writers[0].Schema
	grammar := driver.Grammar()
	processor := driver.Processor()

	return &Schema{
		config:     config,
		driver:     driver,
		grammar:    grammar,
		log:        log,
		migrations: migrations,
		orm:        orm,
		prefix:     prefix,
		processor:  processor,
		schema:     schema,
		goTypes:    defaultGoTypes(),
	}, nil
}

func (r *Schema) Connection(name string) contractsschema.Schema {
	schema, err := NewSchema(r.config, r.log, r.orm.Connection(name), r.driver, r.migrations)
	if err != nil {
		r.log.Panic(errors.SchemaConnectionNotFound.Args(name).SetModule(errors.ModuleSchedule).Error())
		return nil
	}

	return schema
}

func (r *Schema) Create(table string, callback func(table contractsschema.Blueprint)) error {
	blueprint := r.createBlueprint(table)
	blueprint.Create()
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToCreateTable.Args(table, err)
	}

	return nil
}

func (r *Schema) Drop(table string) error {
	blueprint := r.createBlueprint(table)
	blueprint.Drop()

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToDropTable.Args(table, err)
	}

	return nil
}

func (r *Schema) DropAllTables() error {
	tables, err := r.GetTables()
	if err != nil {
		return err
	}

	sqls := r.grammar.CompileDropAllTables(r.schema, tables)
	if sqls == nil {
		return nil
	}

	return r.orm.Transaction(func(tx contractsorm.Query) error {
		for _, sql := range sqls {
			if _, err := tx.Exec(sql); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Schema) DropAllTypes() error {
	types, err := r.GetTypes()
	if err != nil {
		return err
	}

	return r.orm.Transaction(func(tx contractsorm.Query) error {
		for _, sql := range r.grammar.CompileDropAllTypes(r.schema, types) {
			if _, err := tx.Exec(sql); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Schema) DropAllViews() error {
	views, err := r.GetViews()
	if err != nil {
		return err
	}

	sqls := r.grammar.CompileDropAllViews(r.schema, views)
	if sqls == nil {
		return nil
	}

	return r.orm.Transaction(func(tx contractsorm.Query) error {
		for _, sql := range sqls {
			if _, err := tx.Exec(sql); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Schema) DropColumns(table string, columns []string) error {
	blueprint := r.createBlueprint(table)
	blueprint.DropColumn(columns...)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToDropColumns.Args(table, err)
	}

	return nil
}

func (r *Schema) DropIfExists(table string) error {
	blueprint := r.createBlueprint(table)
	blueprint.DropIfExists()

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToDropTable.Args(table, err)
	}

	return nil
}

func (r *Schema) Extend(extend *contractsschema.Extension) contractsschema.Schema {
	r.extendGoTypes(extend.GoTypes)
	return r
}

func (r *Schema) GetColumnListing(table string) []string {
	columns, err := r.GetColumns(table)
	if err != nil {
		r.log.Errorf("failed to get %s columns: %v", table, err)
		return nil
	}

	var names []string
	for _, column := range columns {
		names = append(names, column.Name)
	}

	return names
}

func (r *Schema) GetColumns(table string) ([]driver.Column, error) {
	var dbColumns []driver.DBColumn
	sql, err := r.grammar.CompileColumns(r.schema, table)
	if err != nil {
		return nil, err
	}

	if err := r.orm.Query().Raw(sql).Scan(&dbColumns); err != nil {
		return nil, err
	}

	return r.processor.ProcessColumns(dbColumns), nil
}

func (r *Schema) GetConnection() string {
	return r.orm.Name()
}

func (r *Schema) GetForeignKeys(table string) ([]driver.ForeignKey, error) {
	table = r.prefix + table

	var dbForeignKeys []driver.DBForeignKey
	if err := r.orm.Query().Raw(r.grammar.CompileForeignKeys(r.schema, table)).Scan(&dbForeignKeys); err != nil {
		return nil, err
	}

	return r.processor.ProcessForeignKeys(dbForeignKeys), nil
}

func (r *Schema) GetIndexListing(table string) []string {
	indexes, err := r.GetIndexes(table)
	if err != nil {
		r.log.Errorf("failed to get %s indexes: %v", table, err)
		return nil
	}

	var names []string
	for _, index := range indexes {
		names = append(names, index.Name)
	}

	return names
}

func (r *Schema) GetIndexes(table string) ([]driver.Index, error) {
	var dbIndexes []driver.DBIndex
	sql, err := r.grammar.CompileIndexes(r.schema, table)
	if err != nil {
		return nil, err
	}

	if err := r.orm.Query().Raw(sql).Scan(&dbIndexes); err != nil {
		return nil, err
	}

	return r.processor.ProcessIndexes(dbIndexes), nil
}

func (r *Schema) GetTableListing() []string {
	tables, err := r.GetTables()
	if err != nil {
		r.log.Errorf("failed to get tables: %v", err)
		return nil
	}

	var names []string
	for _, table := range tables {
		names = append(names, table.Name)
	}

	return names
}

func (r *Schema) GetTables() ([]driver.Table, error) {
	var tables []driver.Table
	if err := r.orm.Query().Raw(r.grammar.CompileTables(r.orm.DatabaseName())).Scan(&tables); err != nil {
		return nil, err
	}

	return tables, nil
}

func (r *Schema) GetTypes() ([]driver.Type, error) {
	var types []driver.Type
	if err := r.orm.Query().Raw(r.grammar.CompileTypes()).Scan(&types); err != nil {
		return nil, err
	}

	return r.processor.ProcessTypes(types), nil
}

func (r *Schema) GetViews() ([]driver.View, error) {
	var views []driver.View
	if err := r.orm.Query().Raw(r.grammar.CompileViews(r.orm.DatabaseName())).Scan(&views); err != nil {
		return nil, err
	}

	return views, nil
}

func (r *Schema) GoTypes() []contractsschema.GoType {
	return r.goTypes
}

func (r *Schema) HasColumn(table, column string) bool {
	return slices.Contains(r.GetColumnListing(table), column)
}

func (r *Schema) HasColumns(table string, columns []string) bool {
	columnListing := r.GetColumnListing(table)
	for _, column := range columns {
		if !slices.Contains(columnListing, column) {
			return false
		}
	}

	return true
}

func (r *Schema) HasIndex(table, index string) bool {
	indexListing := r.GetIndexListing(table)

	return slices.Contains(indexListing, index)
}

func (r *Schema) HasTable(name string) bool {
	var schema string
	if strings.Contains(name, ".") {
		lastDotIndex := strings.LastIndex(name, ".")
		schema = name[:lastDotIndex]
		name = name[lastDotIndex+1:]
	}

	tableName := r.prefix + name

	tables, err := r.GetTables()
	if err != nil {
		r.log.Errorf(errors.SchemaFailedToGetTables.Args(r.orm.Name(), err).Error())
		return false
	}

	for _, table := range tables {
		if table.Name == tableName {
			if schema == "" || schema == table.Schema {
				return true
			}
		}
	}

	return false
}

func (r *Schema) HasType(name string) bool {
	types, err := r.GetTypes()
	if err != nil {
		r.log.Errorf(errors.SchemaFailedToGetTables.Args(r.orm.Name(), err).Error())
		return false
	}

	for _, t := range types {
		if t.Name == name {
			return true
		}
	}

	return false
}

func (r *Schema) HasView(name string) bool {
	views, err := r.GetViews()
	if err != nil {
		r.log.Errorf(errors.SchemaFailedToGetTables.Args(r.orm.Name(), err).Error())
		return false
	}

	for _, view := range views {
		if view.Name == name {
			return true
		}
	}

	return false
}

func (r *Schema) Migrations() []contractsschema.Migration {
	return r.migrations
}

func (r *Schema) Orm() contractsorm.Orm {
	return r.orm
}

func (r *Schema) Register(migrations []contractsschema.Migration) {
	r.migrations = migrations
}

func (r *Schema) Rename(from, to string) error {
	blueprint := r.createBlueprint(from)
	blueprint.Rename(to)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToRenameTable.Args(from, err)
	}

	return nil
}

func (r *Schema) SetConnection(name string) {
	r.orm = r.orm.Connection(name)
}

func (r *Schema) Sql(sql string) error {
	_, err := r.orm.Query().Exec(sql)

	return err
}

func (r *Schema) Table(table string, callback func(table contractsschema.Blueprint)) error {
	blueprint := r.createBlueprint(table)
	callback(blueprint)

	if err := r.build(blueprint); err != nil {
		return errors.SchemaFailedToChangeTable.Args(table, err)
	}

	return nil
}

func (r *Schema) build(blueprint contractsschema.Blueprint) error {
	if r.orm.Query().InTransaction() {
		return blueprint.Build(r.orm.Query(), r.grammar)
	}

	return r.orm.Transaction(func(tx contractsorm.Query) error {
		return blueprint.Build(tx, r.grammar)
	})
}

func (r *Schema) createBlueprint(table string) contractsschema.Blueprint {
	return NewBlueprint(r, r.prefix, table)
}

func (r *Schema) extendGoTypes(goTypes []contractsschema.GoType) {
	if len(goTypes) == 0 {
		return
	}

	defaults := r.goTypes
	fallbackIdx := len(defaults) - 1
	for i, m := range defaults {
		if m.Pattern == ".*" {
			fallbackIdx = i
			break
		}
	}

	patternMap := make(map[string]int, len(defaults))
	for i, m := range defaults {
		patternMap[m.Pattern] = i
	}

	newPatternCount := 0
	for _, cfg := range goTypes {
		if _, exists := patternMap[cfg.Pattern]; !exists {
			newPatternCount++
		}
	}

	var result []contractsschema.GoType
	if newPatternCount == 0 {
		result = make([]contractsschema.GoType, len(defaults))
		copy(result, defaults)

		for _, cfg := range goTypes {
			if idx, exists := patternMap[cfg.Pattern]; exists {
				if cfg.Type != "" {
					result[idx].Type = cfg.Type
				}
				if cfg.NullType != "" {
					result[idx].NullType = cfg.NullType
				}
				if cfg.Imports != nil {
					result[idx].Imports = cfg.Imports
				}
			}
		}

		r.goTypes = result
		return
	}

	result = make([]contractsschema.GoType, 0, len(defaults)+newPatternCount)

	result = append(result, defaults[:fallbackIdx]...)

	for _, cfg := range goTypes {
		if _, exists := patternMap[cfg.Pattern]; !exists {
			result = append(result, cfg)
		}
	}

	result = append(result, defaults[fallbackIdx:]...)

	for _, cfg := range goTypes {
		if _, exists := patternMap[cfg.Pattern]; exists {
			for i, mapping := range result {
				if mapping.Pattern == cfg.Pattern {
					if cfg.Type != "" {
						result[i].Type = cfg.Type
					}
					if cfg.NullType != "" {
						result[i].NullType = cfg.NullType
					}
					if cfg.Imports != nil {
						result[i].Imports = cfg.Imports
					}
					break
				}
			}
		}
	}

	r.goTypes = result
}

func defaultGoTypes() []contractsschema.GoType {
	return []contractsschema.GoType{
		// Special cases first - these need to be matched before general patterns
		{Pattern: "(?i)^tinyint\\(1\\)$", Type: "bool", NullType: "*bool"}, // MySQL boolean representation

		// Boolean types
		{Pattern: "(?i)^bool$", Type: "bool", NullType: "*bool"},
		{Pattern: "(?i)^boolean$", Type: "bool", NullType: "*bool"},
		{Pattern: "(?i)^bit\\(1\\)$", Type: "bool", NullType: "*bool"}, // Single bit as boolean
		{Pattern: "(?i)^bit$", Type: "bool", NullType: "*bool"},

		// Integer types - ordered from most specific to general
		{Pattern: "(?i)^bigserial$", Type: "int64", NullType: "*int64"}, // PostgreSQL
		{Pattern: "(?i)^bigint$", Type: "int64", NullType: "*int64"},
		{Pattern: "(?i)^smallserial$", Type: "int16", NullType: "*int16"}, // PostgreSQL
		{Pattern: "(?i)^smallint$", Type: "int16", NullType: "*int16"},
		{Pattern: "(?i)^int2$", Type: "int16", NullType: "*int16"}, // PostgreSQL
		{Pattern: "(?i)^serial$", Type: "int", NullType: "*int"},   // PostgreSQL
		{Pattern: "(?i)^integer$", Type: "int", NullType: "*int"},
		{Pattern: "(?i)^int$", Type: "int", NullType: "*int"},
		{Pattern: "(?i)^int4$", Type: "int", NullType: "*int"},          // PostgreSQL
		{Pattern: "(?i)^mediumint$", Type: "int32", NullType: "*int32"}, // MySQL
		{Pattern: "(?i)^tinyint$", Type: "int8", NullType: "*int8"},     // MySQL (when not tinyint(1))
		{Pattern: "(?i)^int8$", Type: "int8", NullType: "*int8"},        // PostgreSQL

		// Floating point types
		{Pattern: "(?i)^double precision$", Type: "float64", NullType: "*float64"},
		{Pattern: "(?i)^double$", Type: "float64", NullType: "*float64"},
		{Pattern: "(?i)^float8$", Type: "float64", NullType: "*float64"}, // PostgreSQL
		{Pattern: "(?i)^float4$", Type: "float32", NullType: "*float32"}, // PostgreSQL
		{Pattern: "(?i)^float$", Type: "float32", NullType: "*float32"},  // MySQL
		{Pattern: "(?i)^real$", Type: "float32", NullType: "*float32"},

		// Decimal types
		{Pattern: "(?i)^money$", Type: "float64", NullType: "*float64"}, // PostgreSQL
		{Pattern: "(?i)^decimal$", Type: "float64", NullType: "*float64"},
		{Pattern: "(?i)^numeric$", Type: "float64", NullType: "*float64"},

		// String types - longer/specific types first
		{Pattern: "(?i)^character varying$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^varchar$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^character$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^longtext$", Type: "string", NullType: "*string"},   // MySQL
		{Pattern: "(?i)^mediumtext$", Type: "string", NullType: "*string"}, // MySQL
		{Pattern: "(?i)^tinytext$", Type: "string", NullType: "*string"},   // MySQL
		{Pattern: "(?i)^nvarchar$", Type: "string", NullType: "*string"},   // SQL Server
		{Pattern: "(?i)^ntext$", Type: "string", NullType: "*string"},      // SQL Server
		{Pattern: "(?i)^nchar$", Type: "string", NullType: "*string"},      // SQL Server
		{Pattern: "(?i)^text$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^char$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^citext$", Type: "string", NullType: "*string"}, // PostgreSQL

		// JSON types
		{Pattern: "(?i)^jsonb$", Type: "string", NullType: "*string"}, // PostgreSQL
		{Pattern: "(?i)^json$", Type: "string", NullType: "*string"},

		// Date and Time types
		{Pattern: "(?i)^timestamptz$", Type: "carbon.DateTime", NullType: "*carbon.DateTime", Imports: []string{"github.com/goravel/framework/support/carbon"}}, // PostgreSQL
		{Pattern: "(?i)^timestamp$", Type: "carbon.DateTime", NullType: "*carbon.DateTime", Imports: []string{"github.com/goravel/framework/support/carbon"}},
		{Pattern: "(?i)^datetime$", Type: "carbon.DateTime", NullType: "*carbon.DateTime", Imports: []string{"github.com/goravel/framework/support/carbon"}}, // MySQL
		{Pattern: "(?i)^timetz$", Type: "carbon.DateTime", NullType: "*carbon.DateTime", Imports: []string{"github.com/goravel/framework/support/carbon"}},   // PostgreSQL
		{Pattern: "(?i)^time$", Type: "carbon.DateTime", NullType: "*carbon.DateTime", Imports: []string{"github.com/goravel/framework/support/carbon"}},
		{Pattern: "(?i)^date$", Type: "carbon.DateTime", NullType: "*carbon.DateTime", Imports: []string{"github.com/goravel/framework/support/carbon"}},
		{Pattern: "(?i)^interval$", Type: "string", NullType: "*string"}, // PostgreSQL

		// Enum types
		{Pattern: "(?i)^enum$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^set$", Type: "string", NullType: "*string"},

		// Binary types - larger types first
		{Pattern: "(?i)^longblob$", Type: "[]byte", NullType: "[]byte"},   // MySQL
		{Pattern: "(?i)^mediumblob$", Type: "[]byte", NullType: "[]byte"}, // MySQL
		{Pattern: "(?i)^tinyblob$", Type: "[]byte", NullType: "[]byte"},   // MySQL
		{Pattern: "(?i)^blob$", Type: "[]byte", NullType: "[]byte"},       // MySQL
		{Pattern: "(?i)^varbinary$", Type: "[]byte", NullType: "[]byte"},  // MySQL/SQL Server
		{Pattern: "(?i)^binary$", Type: "[]byte", NullType: "[]byte"},     // MySQL/SQL Server
		{Pattern: "(?i)^bytea$", Type: "[]byte", NullType: "[]byte"},      // PostgreSQL

		// Network types (PostgreSQL)
		{Pattern: "(?i)^macaddr$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^cidr$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^inet$", Type: "string", NullType: "*string"},

		// Geometric types (PostgreSQL)
		{Pattern: "(?i)^circle$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^polygon$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^path$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^box$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^lseg$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^line$", Type: "string", NullType: "*string"},
		{Pattern: "(?i)^point$", Type: "string", NullType: "*string"},

		// UUID type
		{Pattern: "(?i)^uuid$", Type: "string", NullType: "*string"},

		// XML
		{Pattern: "(?i)^xml$", Type: "string", NullType: "*string"},

		// SQLite specific
		{Pattern: "(?i)^rowid$", Type: "int64", NullType: "*int64"}, // SQLite

		// Fallback for unknown types
		{Pattern: ".*", Type: "any", NullType: "any"},
	}
}
