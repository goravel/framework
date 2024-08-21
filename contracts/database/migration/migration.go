package migration

const (
	DriverDefault = "default"
	DriverSql     = "sql"
)

type Driver interface {
	Create(name string) error
}
