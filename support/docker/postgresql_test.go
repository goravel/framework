package docker

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/orm"
	contractstesting "github.com/goravel/framework/contracts/testing"
	configmocks "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/env"
)

type PostgresqlTestSuite struct {
	suite.Suite
	mockConfig *configmocks.Config
	postgresql *Postgresql
}

func TestPostgresqlTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping tests of using docker")
	}

	suite.Run(t, new(PostgresqlTestSuite))
}

func (s *PostgresqlTestSuite) SetupTest() {
	s.mockConfig = &configmocks.Config{}
	s.postgresql = NewPostgresql("goravel", "goravel", "goravel")
}

func (s *PostgresqlTestSuite) TestBuild() {
	s.Nil(s.postgresql.Build())
	instance, err := s.postgresql.connect()
	s.Nil(err)
	s.NotNil(instance)

	s.Equal("127.0.0.1", s.postgresql.Config().Host)
	s.Equal("goravel", s.postgresql.Config().Database)
	s.Equal("goravel", s.postgresql.Config().Username)
	s.Equal("goravel", s.postgresql.Config().Password)
	s.True(s.postgresql.Config().Port > 0)

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

	s.Nil(s.postgresql.Fresh())

	res = instance.Raw(`
		SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' and table_name = 'users';
		`).Scan(&count)
	s.Nil(res.Error)
	s.Equal(int64(0), count)

	s.Nil(s.postgresql.Stop())
}

func (s *PostgresqlTestSuite) TestImage() {
	image := contractstesting.Image{
		Repository: "postgresql",
	}
	s.postgresql.Image(image)
	s.Equal(&image, s.postgresql.image)
}

func (s *PostgresqlTestSuite) TestName() {
	s.Equal(orm.DriverPostgresql, s.postgresql.Name())
}
