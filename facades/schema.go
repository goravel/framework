package facades

import (
	"github.com/goravel/framework/contracts/database/migration"
)

func Schema() migration.Schema {
	return App().MakeSchema()
}
