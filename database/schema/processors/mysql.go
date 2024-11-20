package processors

import (
	"github.com/goravel/framework/contracts/database/schema"
)

type Mysql struct {
}

func NewMysql() Mysql {
	return Mysql{}
}

func (r Mysql) ProcessColumns(dbColumns []DBColumn) []schema.Column {
	var columns []schema.Column
	for _, dbColumn := range dbColumns {
		var nullable bool
		if dbColumn.Nullable == "YES" {
			nullable = true
		}
		var autoIncrement bool
		if dbColumn.Extra == "auto_increment" {
			autoIncrement = true
		}

		columns = append(columns, schema.Column{
			Autoincrement: autoIncrement,
			Collation:     dbColumn.Collation,
			Comment:       dbColumn.Comment,
			Default:       dbColumn.Default,
			Name:          dbColumn.Name,
			Nullable:      nullable,
			Type:          dbColumn.Type,
			TypeName:      dbColumn.TypeName,
		})
	}

	return columns
}

func (r Mysql) ProcessIndexes(dbIndexes []DBIndex) []schema.Index {
	return processIndexes(dbIndexes)
}
