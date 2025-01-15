package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
)

type OrmSuite struct {
	suite.Suite
	orm           *orm.Orm
	defaultDriver string
	queries       map[string]*TestQuery
}

func TestOrmSuite(t *testing.T) {
	suite.Run(t, &OrmSuite{
		queries: make(map[string]*TestQuery),
	})
}

func (s *OrmSuite) SetupSuite() {
	postgresTestQuery := postgresTestQuery("", false)
	s.queries[postgresTestQuery.Driver().Config().Connection] = postgresTestQuery
	s.defaultDriver = postgresTestQuery.Driver().Config().Driver
}

func (s *OrmSuite) SetupTest() {
	queries := make(map[string]contractsorm.Query)

	for driver, query := range s.queries {
		queries[driver] = query.Query()
	}

	s.orm = orm.NewOrm(context.Background(), nil, s.defaultDriver, queries[s.defaultDriver], queries, nil, nil, nil)
}

func (s *OrmSuite) TearDownSuite() {
	// TODO Shutdown Sqlite
	// if s.queries[database.DriverSqlite] != nil {
	// 	s.NoError(s.queries[database.DriverSqlite].Docker().Shutdown())
	// }
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
