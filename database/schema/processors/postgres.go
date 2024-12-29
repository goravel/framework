package processors

import (
	"strings"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/str"
)

type Postgres struct {
}

func NewPostgres() Postgres {
	return Postgres{}
}

func (r Postgres) ProcessColumns(dbColumns []schema.DBColumn) []schema.Column {
	var columns []schema.Column
	for _, dbColumn := range dbColumns {
		var autoincrement bool
		if str.Of(dbColumn.Default).StartsWith("nextval(") {
			autoincrement = true
		}

		columns = append(columns, schema.Column{
			Autoincrement: autoincrement,
			Collation:     dbColumn.Collation,
			Comment:       dbColumn.Comment,
			Default:       dbColumn.Default,
			Name:          dbColumn.Name,
			Nullable:      cast.ToBool(dbColumn.Nullable),
			Type:          dbColumn.Type,
			TypeName:      dbColumn.TypeName,
		})
	}

	return columns
}

func (r Postgres) ProcessForeignKeys(dbForeignKeys []schema.DBForeignKey) []schema.ForeignKey {
	var foreignKeys []schema.ForeignKey

	short := map[string]string{
		"a": "no action",
		"c": "cascade",
		"d": "set default",
		"n": "set null",
		"r": "restrict",
	}

	for _, dbForeignKey := range dbForeignKeys {
		onUpdate := short[strings.ToLower(dbForeignKey.OnUpdate)]
		if onUpdate == "" {
			onUpdate = strings.ToLower(dbForeignKey.OnUpdate)
		}
		onDelete := short[strings.ToLower(dbForeignKey.OnDelete)]
		if onDelete == "" {
			onDelete = strings.ToLower(dbForeignKey.OnDelete)
		}

		foreignKeys = append(foreignKeys, schema.ForeignKey{
			Name:           dbForeignKey.Name,
			Columns:        strings.Split(dbForeignKey.Columns, ","),
			ForeignSchema:  dbForeignKey.ForeignSchema,
			ForeignTable:   dbForeignKey.ForeignTable,
			ForeignColumns: strings.Split(dbForeignKey.ForeignColumns, ","),
			OnUpdate:       onUpdate,
			OnDelete:       onDelete,
		})
	}

	return foreignKeys
}

func (r Postgres) ProcessIndexes(dbIndexes []schema.DBIndex) []schema.Index {
	return processIndexes(dbIndexes)
}

func (r Postgres) ProcessTypes(types []schema.Type) []schema.Type {
	processType := map[string]string{
		"b": "base",
		"c": "composite",
		"d": "domain",
		"e": "enum",
		"p": "pseudo",
		"r": "range",
		"m": "multirange",
	}
	processCategory := map[string]string{
		"a": "array",
		"b": "boolean",
		"c": "composite",
		"d": "date_time",
		"e": "enum",
		"g": "geometric",
		"i": "network_address",
		"n": "numeric",
		"p": "pseudo",
		"r": "range",
		"s": "string",
		"t": "timespan",
		"u": "user_defined",
		"v": "bit_string",
		"x": "unknown",
		"z": "internal_use",
	}

	for i, t := range types {
		types[i].Type = processType[t.Type]
		types[i].Category = processCategory[t.Category]
	}

	return types
}
