package processors

import (
	"github.com/goravel/framework/contracts/database/schema"
)

type Mysql struct {
}

func NewMysql() Mysql {
	return Mysql{}
}

func (r Mysql) ProcessIndexes(dbIndexes []DBIndex) []schema.Index {
	return processIndexes(dbIndexes)
}
