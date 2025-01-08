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

type Config1 struct {
	Connection string
	Driver     string
	Prefix     string
}

// Config Used in config/database.go
type Config struct {
	Dsn      string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	// Only for Postgres
	Schema string
}

// FullConfig Fill the default value for Config
type FullConfig struct {
	Config
	Driver       Driver
	Connection   string
	Prefix       string
	Singular     bool
	Charset      string // Mysql, Sqlserver
	Loc          string // Mysql
	Sslmode      string // Postgres
	Timezone     string // Postgres
	NoLowerCase  bool
	NameReplacer Replacer
}

type ConfigBuilder interface {
	Reads() []FullConfig
	Writes() []FullConfig
}

// Replacer replacer interface like strings.Replacer
type Replacer interface {
	Replace(name string) string
}
