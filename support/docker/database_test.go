package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
)

type DatabaseTestSuite struct {
	suite.Suite
	database *Database
}

func TestDatabaseTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, new(DatabaseTestSuite))
}

func (s *DatabaseTestSuite) SetupTest() {
}

func (s *DatabaseTestSuite) TestFreshMysql() {
	database, err := InitDatabase()
	s.Nil(err)
	s.NotNil(database)

	instance, err := database.connect(contractsorm.DriverMysql)
	s.Nil(err)
	s.NotNil(instance)

	res := instance.Exec(`
CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  PRIMARY KEY (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`)
	s.Nil(res.Error)

	res = instance.Exec(`
INSERT INTO users (name) VALUES ('goravel');
`)
	s.Nil(res.Error)
	s.Equal(int64(1), res.RowsAffected)

	var count int64
	res = instance.Raw(fmt.Sprintf("SELECT count(*) FROM information_schema.tables WHERE table_schema = '%s' and table_name = 'users';", database.Database)).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(1), count)

	s.Nil(database.freshMysql())

	res = instance.Raw(fmt.Sprintf("SELECT count(*) FROM information_schema.tables WHERE table_schema = '%s' and table_name = 'users';", database.Database)).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(0), count)

	s.Nil(database.Stop())
}

func (s *DatabaseTestSuite) TestFreshPostgresql() {
	database, err := InitDatabase()
	s.Nil(err)
	s.NotNil(database)

	instance, err := database.connect(contractsorm.DriverPostgresql)
	s.Nil(err)
	s.NotNil(instance)

	res := instance.Exec(`
CREATE TABLE users (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL
);
`)
	s.Nil(res.Error)

	res = instance.Exec(`
INSERT INTO users (name) VALUES ('goravel');
`)
	s.Nil(res.Error)
	s.Equal(int64(1), res.RowsAffected)

	var count int64
	res = instance.Raw(`
SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' and table_name = 'users';
`).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(1), count)

	s.Nil(database.freshPostgresql())

	res = instance.Raw(`
	SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' and table_name = 'users';
	`).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(0), count)

	s.Nil(database.Stop())
}

func (s *DatabaseTestSuite) TestFreshSqlserver() {
	database, err := InitDatabase()
	s.Nil(err)
	s.NotNil(database)

	instance, err := database.connect(contractsorm.DriverSqlserver)
	s.Nil(err)
	s.NotNil(instance)

	res := instance.Exec(`
CREATE TABLE users (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  PRIMARY KEY (id)
);
`)
	s.Nil(res.Error)

	res = instance.Exec(`
INSERT INTO users (name) VALUES ('goravel');
`)
	s.Nil(res.Error)
	s.Equal(int64(1), res.RowsAffected)

	var count int64
	res = instance.Raw(`
SELECT count(*) FROM sys.tables WHERE name = 'users';
`).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(1), count)

	s.Nil(database.freshSqlserver())

	res = instance.Raw(`
SELECT count(*) FROM sys.tables WHERE name = 'users';
`).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(0), count)

	s.Nil(database.Stop())
}

func TestInitDatabase(t *testing.T) {
	database1, err := InitDatabase()
	assert.Nil(t, err)
	assert.NotNil(t, database1)
	assert.Equal(t, mysqlPort, database1.MysqlPort)
	assert.Equal(t, postgresqlPort, database1.PostgresqlPort)
	assert.Equal(t, sqlserverPort, database1.SqlserverPort)

	database2, err := InitDatabase()
	assert.Nil(t, err)
	assert.NotNil(t, database2)
	assert.Equal(t, mysqlPort+1, database2.MysqlPort)
	assert.Equal(t, postgresqlPort+1, database2.PostgresqlPort)
	assert.Equal(t, sqlserverPort+1, database2.SqlserverPort)

	mysql1, err := database1.connect(contractsorm.DriverMysql)
	assert.Nil(t, err)
	assert.NotNil(t, mysql1)

	mysql2, err := database2.connect(contractsorm.DriverMysql)
	assert.Nil(t, err)
	assert.NotNil(t, mysql2)

	err = database1.Stop()
	if err != nil {
		fmt.Println("hwb1-------", err.Error())
	}
	assert.Nil(t, err)

	err = database2.Stop()
	if err != nil {
		fmt.Println("hwb2-------", err.Error())
	}
	assert.Nil(t, err)
}
