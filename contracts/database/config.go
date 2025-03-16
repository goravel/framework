package database

import "gorm.io/gorm"

type Pool struct {
	Readers []Config
	Writers []Config
}

type Config struct {
	Connection        string
	Dsn               string
	Database          string
	Dialector         gorm.Dialector
	Driver            string
	Host              string
	NameReplacer      Replacer
	NoLowerCase       bool
	Password          string
	Port              int
	Prefix            string
	Schema            string
	Username          string
	Version           string
	PlaceholderFormat PlaceholderFormat
	Singular          bool
	Sslmode           string
	Timezone          string
}

type PlaceholderFormat interface {
	ReplacePlaceholders(sql string) (string, error)
}

// Replacer replacer interface like strings.Replacer
type Replacer interface {
	Replace(name string) string
}
