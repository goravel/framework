package docker

import (
	"fmt"
)

const (
	DbPassword     = "Goravel(!)"
	DbUser         = "goravel"
	DbDatabase     = "goravel"
	MysqlPort      = 9910
	PostgresqlPort = 9920
	SqlserverPort  = 9930
)

type Compose struct {
}

func (r Compose) Database() string {
	return fmt.Sprintf(`version: '3'

services:
  mysql:
    image: 'mysql:latest'
    ports:
      - %d:3306
    environment:
      - MYSQL_DATABASE=%s
      - MYSQL_USER=%s
      - MYSQL_PASSWORD=%s
      - MYSQL_RANDOM_ROOT_PASSWORD="yes"
  postgresql:
    image: 'postgres:latest'
    ports:
      - %d:5432
    environment:
      - TZ=Asia/Shanghai
      - POSTGRES_DB=%s
      - POSTGRES_USER=%s
      - POSTGRES_PASSWORD=%s
  sqlserver:
    image: 'mcmoe/mssqldocker:latest'
    ports:
      - %d:1433
    environment:
      - ACCEPT_EULA=Y
      - MSSQL_DB=%s
      - MSSQL_USER=%s
      - MSSQL_PASSWORD=%s
      - SA_PASSWORD=%s
`, MysqlPort, DbDatabase, DbUser, DbPassword, PostgresqlPort, DbDatabase, DbUser, DbPassword, SqlserverPort, DbDatabase, DbUser, DbPassword, DbPassword)
}
