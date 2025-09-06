package tests

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	databasedb "github.com/goravel/framework/database/db"
	"github.com/goravel/framework/database/orm"
	"github.com/goravel/postgres"
	"github.com/goravel/sqlite"
)

type OrmSuite struct {
	suite.Suite
	orm               *orm.Orm
	defaultConnection string
	queries           map[string]*TestQuery
}

func TestOrmSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &OrmSuite{
		queries: make(map[string]*TestQuery),
	})
}

func (s *OrmSuite) SetupSuite() {
	s.defaultConnection = postgres.Name
	s.queries = NewTestQueryBuilder().All("", false)
}

func (s *OrmSuite) SetupTest() {
	queries := make(map[string]contractsorm.Query)

	for driver, query := range s.queries {
		query.CreateTable(TestTableRoles)
		queries[driver] = query.Query()
	}

	dbConfig := s.queries[s.defaultConnection].Driver().Pool().Writers[0]
	s.orm = orm.NewOrm(context.Background(), nil, dbConfig.Connection, dbConfig, queries[s.defaultConnection], queries, nil, nil, nil)
}

func (s *OrmSuite) TearDownSuite() {
	if s.queries[sqlite.Name] != nil {
		docker, err := s.queries[sqlite.Name].Driver().Docker()
		s.NoError(err)
		s.NoError(docker.Shutdown())
	}
}

func (s *OrmSuite) TestConnection() {
	for connection := range s.queries {
		s.NotNil(s.orm.Connection(connection))
	}
}

func (s *OrmSuite) TestDB() {
	db, err := s.orm.DB()
	s.NotNil(db)
	s.Nil(err)

	for connection := range s.queries {
		db, err := s.orm.Connection(connection).DB()
		s.NotNil(db)
		s.Nil(err)
	}
}

func (s *OrmSuite) TestQuery() {
	s.NotNil(s.orm.Query())

	s.NotPanics(func() {
		for i := 0; i < 5; i++ {
			go func() {
				var role Role
				_ = s.orm.Query().Find(&role, 1)
			}()
		}
	})

	for connection := range s.queries {
		s.NotNil(s.orm.Connection(connection).Query())
	}
}

func (s *OrmSuite) TestQueryLog() {
	ctx := databasedb.EnableQueryLog(context.Background())

	role := Role{Name: "query_log_product"}
	s.orm.WithContext(ctx).Query().Create(&role)
	s.True(role.ID > 0)

	var role1 Role
	err := s.orm.WithContext(ctx).Query().Where("name", "query_log_product").First(&role1)
	s.NoError(err)
	s.True(role1.ID > 0)

	queryLogs := databasedb.GetQueryLog(ctx)
	s.Equal(2, len(queryLogs))
	s.Contains(queryLogs[0].Query, "INSERT INTO \"roles\" (\"created_at\",\"updated_at\",\"name\",\"avatar\") VALUES ('")
	s.Contains(queryLogs[0].Query, "'query_log_product','') RETURNING \"id\"")
	s.True(queryLogs[0].Time > 0)
	s.Equal("SELECT * FROM \"roles\" WHERE \"name\" = 'query_log_product' ORDER BY \"roles\".\"id\" LIMIT 1", queryLogs[1].Query)
	s.True(queryLogs[1].Time > 0)

	ctx = databasedb.DisableQueryLog(ctx)

	role2 := Role{Name: "query_log_product2"}
	s.orm.WithContext(ctx).Query().Create(&role2)

	queryLogs = databasedb.GetQueryLog(ctx)
	s.Equal(0, len(queryLogs))
}

func (s *OrmSuite) TestFactory() {
	s.NotNil(s.orm.Factory())

	for connection := range s.queries {
		s.NotNil(s.orm.Connection(connection).Factory())
	}
}

func (s *OrmSuite) TestObserve() {
	s.orm.Observe(Role{}, &UserObserver{})

	for connection := range s.queries {
		role := Role{Name: "observer_name"}
		s.EqualError(s.orm.Connection(connection).Query().Create(&role), "error")
	}
}

func (s *OrmSuite) TestTransactionSuccess() {
	for connection := range s.queries {
		role := Role{Name: "transaction_success_role", Avatar: "transaction_success_avatar"}
		role1 := Role{Name: "transaction_success_role1", Avatar: "transaction_success_avatar1"}
		s.Nil(s.orm.Connection(connection).Transaction(func(tx contractsorm.Query) error {
			s.Nil(tx.Create(&role))
			s.Nil(tx.Create(&role1))

			return nil
		}))

		var role2, role3 Role
		s.Nil(s.orm.Connection(connection).Query().Find(&role2, role.ID))
		s.Nil(s.orm.Connection(connection).Query().Find(&role3, role1.ID))
	}
}

func (s *OrmSuite) TestTransactionError() {
	for connection := range s.queries {
		s.NotNil(s.orm.Connection(connection).Transaction(func(tx contractsorm.Query) error {
			role := Role{Name: "transaction_error_role", Avatar: "transaction_error_avatar"}
			s.Nil(tx.Create(&role))

			role1 := Role{Name: "transaction_error_role1", Avatar: "transaction_error_avatar1"}
			s.Nil(tx.Create(&role1))

			return errors.New("error")
		}))

		var roles []Role
		s.Nil(s.orm.Connection(connection).Query().Find(&roles))
		s.Equal(0, len(roles))
	}
}

func (s *OrmSuite) TestTransactionPanic() {
	for connection := range s.queries {
		err := s.orm.Connection(connection).Transaction(func(tx contractsorm.Query) error {
			role := Role{Name: "transaction_error_role", Avatar: "transaction_error_avatar"}
			s.Nil(tx.Create(&role))

			role1 := Role{Name: "transaction_error_role1", Avatar: "transaction_error_avatar1"}
			s.Nil(tx.Create(&role1))

			panic(1)
		})

		s.Equal(fmt.Errorf("panic: %v", 1), err)

		var roles []Role
		s.Nil(s.orm.Connection(connection).Query().Find(&roles))
		s.Equal(0, len(roles))
	}
}

func (s *OrmSuite) TestWithContext() {
	s.orm.Observe(Role{}, &UserObserver{})
	ctx := context.WithValue(context.Background(), testContextKey, "with_context_goravel")
	role := Role{Name: "with_context_name"}

	// Call Query directly
	err := s.orm.WithContext(ctx).Query().Create(&role)
	s.Nil(err)
	s.Equal("with_context_name", role.Name)
	s.Equal("with_context_goravel", role.Avatar)

	// Call Connection, then call WithContext
	for connection := range s.queries {
		role.ID = 0
		role.Avatar = ""
		err := s.orm.Connection(connection).WithContext(ctx).Query().Create(&role)
		s.Nil(err)
		s.Equal("with_context_name", role.Name)
		s.Equal("with_context_goravel", role.Avatar)
	}

	// Call WithContext, then call Connection
	for connection := range s.queries {
		role.ID = 0
		role.Avatar = ""
		err := s.orm.WithContext(ctx).Connection(connection).Query().Create(&role)
		s.Nil(err)
		s.Equal("with_context_name", role.Name)
		s.Equal("with_context_goravel", role.Avatar)
	}
}
