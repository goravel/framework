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

func (r *Schema) GoTypeMap() map[string]contractsschema.GoTypeMapping {
	typeMapping := getDefaultGoTypeMapping()
	mappingConfig, ok := r.config.Get("database.model.mapping").(map[string]contractsschema.GoTypeMapping)
	if ok {
		for schemaType, configMapping := range mappingConfig {
			if defaultEntry, exists := typeMapping[schemaType]; exists {
				if configMapping.Type != "" {
					defaultEntry.Type = configMapping.Type
				}
				if configMapping.NullType != "" {
					defaultEntry.NullType = configMapping.NullType
				}
				if configMapping.Imports != nil {
					defaultEntry.Imports = configMapping.Imports
				}
				if configMapping.PrecisionBasedTypes != nil {
					defaultEntry.PrecisionBasedTypes = configMapping.PrecisionBasedTypes
				}
				typeMapping[schemaType] = defaultEntry
			} else {
				typeMapping[schemaType] = configMapping
			}
		}
	}

	return typeMapping
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

func getDefaultGoTypeMapping() map[string]contractsschema.GoTypeMapping {
	return map[string]contractsschema.GoTypeMapping{
		contractsschema.TypeBigInteger.Value(): {
			Type:     "int64",
			NullType: "*int64",
			Imports:  []string{},
		},
		contractsschema.TypeBoolean.Value(): {
			Type:     "bool",
			NullType: "*bool",
			Imports:  []string{},
		},
		contractsschema.TypeChar.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
		contractsschema.TypeDate.Value(): {
			Type:     "carbon.DateTime",
			NullType: "*carbon.DateTime",
			Imports:  []string{"github.com/goravel/framework/support/carbon"},
		},
		contractsschema.TypeDateTime.Value(): {
			Type:     "carbon.DateTime",
			NullType: "*carbon.DateTime",
			Imports:  []string{"github.com/goravel/framework/support/carbon"},
		},
		contractsschema.TypeDateTimeTZ.Value(): {
			Type:     "carbon.DateTime",
			NullType: "*carbon.DateTime",
			Imports:  []string{"github.com/goravel/framework/support/carbon"},
		},
		contractsschema.TypeDecimal.Value(): {
			Type:     "float64",
			NullType: "*float64",
			Imports:  []string{},
		},
		contractsschema.TypeDouble.Value(): {
			Type:     "float64",
			NullType: "*float64",
			Imports:  []string{},
		},
		contractsschema.TypeEnum.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
		contractsschema.TypeFloat.Value(): {
			Type:     "float32",
			NullType: "*float32",
			Imports:  []string{},
		},
		contractsschema.TypeInteger.Value(): {
			Type:     "int",
			NullType: "*int",
			Imports:  []string{},
		},
		contractsschema.TypeJson.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
		contractsschema.TypeJsonb.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
		contractsschema.TypeLongText.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
		contractsschema.TypeMediumInteger.Value(): {
			Type:     "int",
			NullType: "*int",
			Imports:  []string{},
		},
		contractsschema.TypeMediumText.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
		contractsschema.TypeSmallInteger.Value(): {
			Type:     "int16",
			NullType: "*int16",
			Imports:  []string{},
		},
		contractsschema.TypeString.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
		contractsschema.TypeText.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
		contractsschema.TypeTime.Value(): {
			Type:     "carbon.DateTime",
			NullType: "*carbon.DateTime",
			Imports:  []string{"github.com/goravel/framework/support/carbon"},
		},
		contractsschema.TypeTimeTZ.Value(): {
			Type:     "carbon.DateTime",
			NullType: "*carbon.DateTime",
			Imports:  []string{"github.com/goravel/framework/support/carbon"},
		},
		contractsschema.TypeTimestamp.Value(): {
			Type:     "carbon.DateTime",
			NullType: "*carbon.DateTime",
			Imports:  []string{"github.com/goravel/framework/support/carbon"},
		},
		contractsschema.TypeTimestampTZ.Value(): {
			Type:     "carbon.DateTime",
			NullType: "*carbon.DateTime",
			Imports:  []string{"github.com/goravel/framework/support/carbon"},
		},
		contractsschema.TypeTinyInteger.Value(): {
			Type:     "int8",
			NullType: "*int8",
			Imports:  []string{},
		},
		contractsschema.TypeTinyText.Value(): {
			Type:     "string",
			NullType: "*string",
			Imports:  []string{},
		},
	}
}
