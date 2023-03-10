package database

type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

type Result struct {
	RowsAffected int64
}
