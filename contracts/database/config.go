package database

type Config struct {
	Connection string
	Database   string
	Driver     string
	Host       string
	Password   string
	Port       int
	Prefix     string
	Schema     string
	Username   string
	Version    string
}
