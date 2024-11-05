package processors

import (
	"strings"

	"github.com/goravel/framework/contracts/database/schema"
)

type Postgres struct {
}

func NewPostgres() Postgres {
	return Postgres{}
}

func (r *Postgres) ProcessIndexes(indexes []schema.Index) []schema.Index {
	for i, index := range indexes {
		indexes[i].Name = strings.ToLower(index.Name)
		indexes[i].Type = strings.ToLower(index.Type)
	}

	return indexes
}

func (r *Postgres) ProcessTypes(types []schema.Type) []schema.Type {
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
