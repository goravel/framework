package database

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

// Config Used in config/database.go
type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// FullConfig Fill the default value for Config
type FullConfig struct {
	Config
	Driver     Driver
	Connection string
	Prefix     string
	Singular   bool
	Charset    string // Mysql, Sqlserver
	Loc        string // Mysql
	Sslmode    string // Postgres
	Timezone   string // Postgres
}

type ConfigBuilder interface {
	Reads() []FullConfig
	Writes() []FullConfig
}
