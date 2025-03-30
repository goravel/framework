package facades

import (
	"github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/contracts/database/orm"
)

func DB() db.DB {
	return App().MakeDB()
}

func Orm() orm.Orm {
	return App().MakeOrm()
}
