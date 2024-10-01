package orm

// DEPRECATED Move to contracts/database/config.go
const (
	DriverMysql     Driver = "mysql"
	DriverPostgres  Driver = "postgres"
	DriverSqlite    Driver = "sqlite"
	DriverSqlserver Driver = "sqlserver"
)

type Driver string

func (d Driver) String() string {
	return string(d)
}
