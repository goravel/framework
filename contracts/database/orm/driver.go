package orm

type Driver string

func (d Driver) String() string {
	return string(d)
}

const (
	DriverMysql      Driver = "mysql"
	DriverPostgresql Driver = "postgresql"
	DriverSqlite     Driver = "sqlite"
	DriverSqlserver  Driver = "sqlserver"
)
