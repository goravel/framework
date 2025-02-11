package database

type Config struct {
	Connection        string
	DNS               string
	Database          string
	Driver            string
	Host              string
	Password          string
	Port              int
	Prefix            string
	Schema            string
	Username          string
	Version           string
	PlaceholderFormat PlaceholderFormat
}

type PlaceholderFormat interface {
	ReplacePlaceholders(sql string) (string, error)
}
