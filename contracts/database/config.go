package database

// TODO Remove this
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

type Config struct {
	Connection string
	Driver     string
	// TODO Check if it can be removed
	Prefix string
	// TODO Check if it can be removed
	Schema string
}
