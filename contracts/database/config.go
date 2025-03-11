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
