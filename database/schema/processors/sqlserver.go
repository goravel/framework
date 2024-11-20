package processors

import (
	"fmt"

	"github.com/spf13/cast"

	"github.com/goravel/framework/contracts/database/schema"
)

type Sqlserver struct {
}

func NewSqlserver() Sqlserver {
	return Sqlserver{}
}

func (r Sqlserver) ProcessColumns(dbColumns []DBColumn) []schema.Column {
	var columns []schema.Column
	for _, dbColumn := range dbColumns {
		columns = append(columns, schema.Column{
			Autoincrement: dbColumn.Autoincrement,
			Collation:     dbColumn.Collation,
			Comment:       dbColumn.Comment,
			Default:       dbColumn.Default,
			Name:          dbColumn.Name,
			Nullable:      cast.ToBool(dbColumn.Nullable),
			Type:          getType(dbColumn),
			TypeName:      dbColumn.TypeName,
		})
	}

	return columns
}

func (r Sqlserver) ProcessIndexes(dbIndexes []DBIndex) []schema.Index {
	return processIndexes(dbIndexes)
}

func getType(dbColumn DBColumn) string {
	var typeName string
	switch dbColumn.TypeName {
	case "binary", "varbinary", "char", "varchar", "nchar", "nvarchar":
		if dbColumn.Length == -1 {
			typeName = dbColumn.TypeName + "(max)"
		} else {
			typeName = fmt.Sprintf("%s(%d)", dbColumn.TypeName, dbColumn.Length)
		}
	case "decimal", "numeric":
		typeName = fmt.Sprintf("%s(%d,%d)", dbColumn.TypeName, dbColumn.Precision, dbColumn.Places)
	case "float", "datetime2", "datetimeoffset", "time":
		typeName = fmt.Sprintf("%s(%d)", dbColumn.TypeName, dbColumn.Precision)
	default:
		typeName = dbColumn.TypeName
	}

	return typeName
}
