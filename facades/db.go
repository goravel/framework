package facades

import (
	"github.com/goravel/framework/contracts/database/orm"
)

var Orm orm.Orm

//Gorm temporary use, will be deleted in the future, please use Orm priority.
//var Gorm *gorm.DB
