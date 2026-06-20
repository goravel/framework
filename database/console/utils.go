package console

import (
	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/errors"
)

func requireSchemaOrm(ctx console.Context, s schema.Schema) bool {
	if orm := s.Orm(); orm == nil || orm.Query() == nil {
		ctx.Error(errors.SchemaOrmNotAvailable.Error())
		return false
	}

	return true
}
