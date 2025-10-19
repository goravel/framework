package main

type Config struct {
	Key         string
	Value       string
	Annotations []string
}

type Stubs struct{}

func (s Stubs) Config() []Config {
	return []Config{
		// // Default database connection name
		// "default": "",
		{
			Key:   "default",
			Value: `""`,
			Annotations: []string{
				"Default database connection name",
			},
		},
		// // Database connections
		// "connections": map[string]any{},
		{
			Key:   "connections",
			Value: `map[string]any{}`,
			Annotations: []string{
				"Database connections",
			},
		},
		// // Pool configuration
		// "pool": map[string]any{
		// 	// Sets the maximum number of connections in the idle
		// 	// connection pool.
		// 	//
		// 	// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
		// 	// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
		// 	//
		// 	// If n <= 0, no idle connections are retained.
		// 	"max_idle_conns": 10,
		// 	// Sets the maximum number of open connections to the database.
		// 	//
		// 	// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
		// 	// MaxIdleConns, then MaxIdleConns will be reduced to match the new
		// 	// MaxOpenConns limit.
		// 	//
		// 	// If n <= 0, then there is no limit on the number of open connections.
		// 	"max_open_conns": 100,
		// 	// Sets the maximum amount of time a connection may be idle.
		// 	//
		// 	// Expired connections may be closed lazily before reuse.
		// 	//
		// 	// If d <= 0, connections are not closed due to a connection's idle time.
		// 	// Unit: Second
		// 	"conn_max_idletime": 3600,
		// 	// Sets the maximum amount of time a connection may be reused.
		// 	//
		// 	// Expired connections may be closed lazily before reuse.
		// 	//
		// 	// If d <= 0, connections are not closed due to a connection's age.
		// 	// Unit: Second
		// 	"conn_max_lifetime": 3600,
		// },
		{
			Key: "pool",
			Value: `map[string]any{
					// Sets the maximum number of connections in the idle
					// connection pool.
					//
					// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
					// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
					//
					// If n <= 0, no idle connections are retained.
					"max_idle_conns": 10,
					// Sets the maximum number of open connections to the database.
					//
					// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
					// MaxIdleConns, then MaxIdleConns will be reduced to match the new
					// MaxOpenConns limit.
					//
					// If n <= 0, then there is no limit on the number of open connections.
					"max_open_conns": 100,
					// Sets the maximum amount of time a connection may be idle.
					//
					// Expired connections may be closed lazily before reuse.
					//
					// If d <= 0, connections are not closed due to a connection's idle time.
					// Unit: Second
					"conn_max_idletime": 3600,
					// Sets the maximum amount of time a connection may be reused.
					//
					// Expired connections may be closed lazily before reuse.
					//
					// If d <= 0, connections are not closed due to a connection's age.
					// Unit: Second
					"conn_max_lifetime": 3600,
		 }`,
			Annotations: []string{
				"Pool configuration",
			},
		},
		// // Sets the threshold for slow queries in milliseconds, the slow query will be logged.
		// // Unit: Millisecond
		// "slow_threshold": 200,
		{
			Key:   "slow_threshold",
			Value: `200`,
			Annotations: []string{
				"Sets the threshold for slow queries in milliseconds, the slow query will be logged.",
				"Unit: Millisecond",
			},
		},
		// // Migration Repository Table
		// //
		// // This table keeps track of all the migrations that have already run for
		// // your application. Using this information, we can determine which of
		// // the migrations on disk haven't actually been run in the database.
		// "migrations": map[string]any{
		// 	"table": "migrations",
		// },
		{
			Key: "migrations",
			Value: `map[string]any{
					"table": "migrations",
				}`,
			Annotations: []string{
				"Migration Repository Table",
				"",
				"This table keeps track of all the migrations that have already run for",
				"your application. Using this information, we can determine which of",
				"the migrations on disk haven't actually been run in the database.",
			},
		},
	}

}

func (s Stubs) DBFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/database/db"
)

func DB() db.DB {
	return App().MakeDB()
}
`
}

func (s Stubs) OrmFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/database/orm"
)

func Orm() orm.Orm {
	return App().MakeOrm()
}
`
}

func (s Stubs) SchemaFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/database/schema"
)

func Schema() schema.Schema {
	return App().MakeSchema()
}
`
}

func (s Stubs) SeederFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/database/seeder"
)

func Seeder() seeder.Facade {
	return App().MakeSeeder()
}
`
}

func (s Stubs) Kernel() string {
	return `package database

import (
	"github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/database/seeder"
)

type Kernel struct {
}

func (kernel Kernel) Migrations() []schema.Migration {
	return []schema.Migration{}
}

func (kernel Kernel) Seeders() []seeder.Seeder {
	return []seeder.Seeder{}
}
`
}
