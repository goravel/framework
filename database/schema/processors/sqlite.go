package processors

import (
	"strings"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/support/collect"
)

type Sqlite struct {
}

func NewSqlite() Sqlite {
	return Sqlite{}
}

func (r Sqlite) ProcessColumns(dbColumns []schema.DBColumn) []schema.Column {
	var primaryKeyNum int
	collect.Map(dbColumns, func(dbColumn schema.DBColumn, _ int) bool {
		if dbColumn.Primary {
			primaryKeyNum++
		}

		return true
	})

	var columns []schema.Column
	for _, dbColumn := range dbColumns {
		ttype := strings.ToLower(dbColumn.Type)
		typeNameParts := strings.SplitN(ttype, "(", 2)
		typeName := ""
		if len(typeNameParts) > 0 {
			typeName = typeNameParts[0]
		}

		columns = append(columns, schema.Column{
			Autoincrement: primaryKeyNum == 1 && dbColumn.Primary && ttype == "integer",
			Default:       dbColumn.Default,
			Name:          dbColumn.Name,
			Nullable:      cast.ToBool(dbColumn.Nullable),
			Type:          ttype,
			TypeName:      typeName,
		})
	}

	return columns
}

func (r Sqlite) ProcessForeignKeys(dbForeignKeys []schema.DBForeignKey) []schema.ForeignKey {
	var foreignKeys []schema.ForeignKey
	for _, dbForeignKey := range dbForeignKeys {
		foreignKeys = append(foreignKeys, schema.ForeignKey{
			Name:           dbForeignKey.Name,
			Columns:        strings.Split(dbForeignKey.Columns, ","),
			ForeignTable:   dbForeignKey.ForeignTable,
			ForeignColumns: strings.Split(dbForeignKey.ForeignColumns, ","),
			OnUpdate:       strings.ToLower(dbForeignKey.OnUpdate),
			OnDelete:       strings.ToLower(dbForeignKey.OnDelete),
		})
	}

	return foreignKeys
}

func (r Sqlite) ProcessIndexes(dbIndexes []schema.DBIndex) []schema.Index {
	var (
		indexes      []schema.Index
		primaryCount int
	)
	for _, dbIndex := range dbIndexes {
		if dbIndex.Primary {
			primaryCount++
		}

		indexes = append(indexes, schema.Index{
			Columns: strings.Split(dbIndex.Columns, ","),
			Name:    strings.ToLower(dbIndex.Name),
			Primary: dbIndex.Primary,
			Unique:  dbIndex.Unique,
		})
	}

	if primaryCount > 1 {
		indexes = collect.Filter(indexes, func(index schema.Index, _ int) bool {
			return !index.Primary
		})
	}

	return indexes
}
