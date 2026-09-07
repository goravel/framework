package db

import (
	contractsdb "github.com/goravel/framework/contracts/database/db"
	"github.com/goravel/framework/database/utils"
)

// Listen registers a listener that is called for every SQL query executed by
// the database, mirroring Laravel's DB::listen. The registry is global, so a
// listener sees queries from every connection.
func Listen(listener func(event *contractsdb.QueryExecuted) error) {
	utils.ListenQuery(listener)
}

// Listen implements contractsdb.DB.
func (r *DB) Listen(listener func(event *contractsdb.QueryExecuted) error) {
	Listen(listener)
}
