package processors

import (
	"strings"

	"github.com/goravel/framework/contracts/database/schema"
)

func processIndexes(dbIndexes []DBIndex) []schema.Index {
	var indexes []schema.Index
	for _, dbIndex := range dbIndexes {
		indexes = append(indexes, schema.Index{
			Columns: strings.Split(dbIndex.Columns, ","),
			Name:    strings.ToLower(dbIndex.Name),
			Type:    strings.ToLower(dbIndex.Type),
			Primary: dbIndex.Primary,
			Unique:  dbIndex.Unique,
		})
	}

	return indexes
}
