package database

import "gorm.io/gorm"

type Pool struct {
	Readers []Config
	Writers []Config
}

type Config struct {
	Connection   string
	Dsn          string
	Database     string
	Dialector    gorm.Dialector
	Driver       string
	Host         string
	NameReplacer Replacer
	NoLowerCase  bool
	Password     string
	Port         int
	Prefix       string
	Schema       string
	Username     string
	Version      string
	Singular     bool
	Sslmode      string
	Timezone     string
}

// Replacer replacer interface like strings.Replacer
type Replacer interface {
	Replace(name string) string
}
