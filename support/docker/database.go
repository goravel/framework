package docker

const (
	password = "Goravel123"
	username = "goravel"
	database = "goravel"
)

type Database struct {
	Mysql      *Mysql
	Mysql1     *Mysql
	Postgresql *Postgresql
	Sqlserver  *Sqlserver
	Sqlite     *Sqlite
}

func InitDatabase() (*Database, error) {
	mysqlDocker := NewMysql(database, username, password)
	if err := mysqlDocker.Build(); err != nil {
		return nil, err
	}

	mysql1Docker := NewMysql(database, username, password)
	if err := mysql1Docker.Build(); err != nil {
		return nil, err
	}

	postgresqlDocker := NewPostgresql(database, username, password)
	if err := postgresqlDocker.Build(); err != nil {
		return nil, err
	}

	sqlserverDocker := NewSqlserver(database, username, password)
	if err := sqlserverDocker.Build(); err != nil {
		return nil, err
	}

	sqliteDocker := NewSqlite(database)
	if err := sqliteDocker.Build(); err != nil {
		return nil, err
	}

	return &Database{
		Mysql:      mysqlDocker,
		Mysql1:     mysql1Docker,
		Postgresql: postgresqlDocker,
		Sqlserver:  sqlserverDocker,
		Sqlite:     sqliteDocker,
	}, nil
}

func (r *Database) Fresh() error {
	if err := r.Mysql.Fresh(); err != nil {
		return err
	}
	if err := r.Postgresql.Fresh(); err != nil {
		return err
	}
	if err := r.Sqlserver.Fresh(); err != nil {
		return err
	}
	if err := r.Sqlite.Fresh(); err != nil {
		return err
	}

	return nil
}

func (r *Database) Stop() error {
	if err := r.Mysql.Stop(); err != nil {
		return err
	}
	if err := r.Postgresql.Stop(); err != nil {
		return err
	}
	if err := r.Sqlserver.Stop(); err != nil {
		return err
	}
	if err := r.Sqlite.Stop(); err != nil {
		return err
	}

	return nil
}
