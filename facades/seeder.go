package facades

import (
	"github.com/goravel/framework/contracts/database"
)

func Seeder() database.Seeder {
	return App().MakeSeeder()
}
