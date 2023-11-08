package docker

import (
	"fmt"
)

type Compose struct {
}

func (r Compose) Database(mysqlPort, postgresqlPort, sqlserverPort int) string {
	return fmt.Sprintf(`version: '3'

services:
  mysql_%d:
    image: 'mysql:latest'
    ports:
      - %d:3306
    environment:
      - MYSQL_DATABASE=%s
      - MYSQL_USER=%s
      - MYSQL_PASSWORD=%s
      - MYSQL_RANDOM_ROOT_PASSWORD="yes"
    networks:
      - custom_network_%d
  postgresql_%d:
    image: 'postgres:latest'
    ports:
      - %d:5432
    environment:
      - TZ=Asia/Shanghai
      - POSTGRES_DB=%s
      - POSTGRES_USER=%s
      - POSTGRES_PASSWORD=%s
    networks:
      - custom_network_%d
  sqlserver_%d:
    image: 'mcmoe/mssqldocker:latest'
    ports:
      - %d:1433
    environment:
      - ACCEPT_EULA=Y
      - MSSQL_DB=%s
      - MSSQL_USER=%s
      - MSSQL_PASSWORD=%s
      - SA_PASSWORD=%s
    networks:
      - custom_network_%d
networks:
  custom_network_%d:
    driver: bridge
`, usingDatabaseNum, mysqlPort, DbDatabase, DbUser, DbPassword, usingDatabaseNum,
		usingDatabaseNum, postgresqlPort, DbDatabase, DbUser, DbPassword, usingDatabaseNum,
		usingDatabaseNum, sqlserverPort, DbDatabase, DbUser, DbPassword, DbPassword, usingDatabaseNum, usingDatabaseNum)
}
