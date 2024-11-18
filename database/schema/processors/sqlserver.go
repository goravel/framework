package processors

import (
	"github.com/goravel/framework/contracts/database/schema"
)

type Sqlserver struct {
}

func NewSqlserver() Sqlserver {
	return Sqlserver{}
}

func (r Sqlserver) ProcessIndexes(dbIndexes []DBIndex) []schema.Index {
	return processIndexes(dbIndexes)
}
