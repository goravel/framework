package orm

const (
	DriverMysql Driver = "mysql"
	// DriverPostgresql DEPRECATED, use DriverPostgres instead.
	DriverPostgresql Driver = "postgresql"
	DriverPostgres   Driver = "postgres"
	DriverSqlite     Driver = "sqlite"
	DriverSqlserver  Driver = "sqlserver"
)

type Driver string

func (d Driver) String() string {
	return string(d)
}
