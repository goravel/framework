package tests

import (
	"testing"

	"github.com/goravel/sqlite"
	"github.com/stretchr/testify/suite"
)

type DBTestSuite struct {
	suite.Suite
	queries map[string]*TestQuery
}

func TestDBTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DBTestSuite{
		queries: make(map[string]*TestQuery),
	})
}

func (s *DBTestSuite) SetupSuite() {
	s.queries = NewTestQueryBuilder().All("", false)
	for _, query := range s.queries {
		query.CreateTable(TestTableUsers)
	}
}

func (s *DBTestSuite) TearDownSuite() {
	if s.queries[sqlite.Name] != nil {
		docker, err := s.queries[sqlite.Name].Driver().Docker()
		s.NoError(err)
		s.NoError(docker.Shutdown())
	}
}

func (s *DBTestSuite) TestWhere() {
	for driver, query := range s.queries {
		s.Run(driver, func() {
			var user []User
			err := query.DB().Table("users").Where("name = ?", "count_user").Get(&user)
			s.NoError(err)
		})
	}
}
