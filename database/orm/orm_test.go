package orm

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/support/env"
)

type contextKey int

const testContextKey contextKey = 0

type User struct {
	Model
	SoftDeletes
	Name   string
	Avatar string
}

type OrmSuite struct {
	suite.Suite
	orm         *Orm
	testQueries map[database.Driver]*gorm.TestQuery
}

func TestOrmSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, &OrmSuite{})
}

func (s *OrmSuite) SetupSuite() {
	s.testQueries = gorm.NewTestQueries().Queries()
	for _, testQuery := range s.testQueries {
		testQuery.CreateTable()
	}
}

func (s *OrmSuite) SetupTest() {
	queries := make(map[string]contractsorm.Query)

	for key, query := range s.testQueries {
		queries[key.String()] = query.Query()
	}

	s.orm = &Orm{
		connection: database.DriverPostgres.String(),
		ctx:        context.Background(),
		query:      queries[database.DriverPostgres.String()],
		queries:    queries,
	}
}

func (s *OrmSuite) TearDownSuite() {
	if s.testQueries[database.DriverSqlite] != nil {
		s.NoError(s.testQueries[database.DriverSqlite].Docker().Shutdown())
	}
}

func (s *OrmSuite) TestConnection() {
	for driver := range s.testQueries {
		s.NotNil(s.orm.Connection(driver.String()))
	}
}

func (s *OrmSuite) TestDB() {
	db, err := s.orm.DB()
	s.NotNil(db)
	s.Nil(err)

	for driver := range s.testQueries {
		db, err := s.orm.Connection(driver.String()).DB()
		s.NotNil(db)
		s.Nil(err)
	}
}

func (s *OrmSuite) TestQuery() {
	s.NotNil(s.orm.Query())

	s.NotPanics(func() {
		for i := 0; i < 5; i++ {
			go func() {
				var user User
				_ = s.orm.Query().Find(&user, 1)
			}()
		}
	})

	for driver := range s.testQueries {
		s.NotNil(s.orm.Connection(driver.String()).Query())
	}
}

func (s *OrmSuite) TestFactory() {
	s.NotNil(s.orm.Factory())

	for driver := range s.testQueries {
		s.NotNil(s.orm.Connection(driver.String()).Factory())
	}
}

func (s *OrmSuite) TestObserve() {
	s.orm.Observe(User{}, &UserObserver{})

	for driver := range s.testQueries {
		user := User{Name: "observer_name"}
		s.EqualError(s.orm.Connection(driver.String()).Query().Create(&user), "error")
	}
}

func (s *OrmSuite) TestTransactionSuccess() {
	for driver := range s.testQueries {
		user := User{Name: "transaction_success_user", Avatar: "transaction_success_avatar"}
		user1 := User{Name: "transaction_success_user1", Avatar: "transaction_success_avatar1"}
		s.Nil(s.orm.Connection(driver.String()).Transaction(func(tx contractsorm.Query) error {
			s.Nil(tx.Create(&user))
			s.Nil(tx.Create(&user1))

			return nil
		}))

		var user2, user3 User
		s.Nil(s.orm.Connection(driver.String()).Query().Find(&user2, user.ID))
		s.Nil(s.orm.Connection(driver.String()).Query().Find(&user3, user1.ID))
	}
}

func (s *OrmSuite) TestTransactionError() {
	for driver := range s.testQueries {
		s.NotNil(s.orm.Connection(driver.String()).Transaction(func(tx contractsorm.Query) error {
			user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
			s.Nil(tx.Create(&user))

			user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
			s.Nil(tx.Create(&user1))

			return errors.New("error")
		}))

		var users []User
		s.Nil(s.orm.Connection(driver.String()).Query().Find(&users))
		s.Equal(0, len(users))
	}
}

func (s *OrmSuite) TestTransactionPanic() {
	for driver := range s.testQueries {
		err := s.orm.Connection(driver.String()).Transaction(func(tx contractsorm.Query) error {
			user := User{Name: "transaction_error_user", Avatar: "transaction_error_avatar"}
			s.Nil(tx.Create(&user))

			user1 := User{Name: "transaction_error_user1", Avatar: "transaction_error_avatar1"}
			s.Nil(tx.Create(&user1))

			panic(1)
		})

		s.Equal(fmt.Errorf("panic: %v", 1), err)

		var users []User
		s.Nil(s.orm.Connection(driver.String()).Query().Find(&users))
		s.Equal(0, len(users))
	}
}

func (s *OrmSuite) TestWithContext() {
	s.orm.Observe(User{}, &UserObserver{})
	ctx := context.WithValue(context.Background(), testContextKey, "with_context_goravel")
	user := User{Name: "with_context_name"}

	// Call Query directly
	err := s.orm.WithContext(ctx).Query().Create(&user)
	s.Nil(err)
	s.Equal("with_context_name", user.Name)
	s.Equal("with_context_goravel", user.Avatar)

	// Call Connection, then call WithContext
	for driver := range s.testQueries {
		user.ID = 0
		user.Avatar = ""
		err := s.orm.Connection(driver.String()).WithContext(ctx).Query().Create(&user)
		s.Nil(err)
		s.Equal("with_context_name", user.Name)
		s.Equal("with_context_goravel", user.Avatar)
	}

	// Call WithContext, then call Connection
	for driver := range s.testQueries {
		user.ID = 0
		user.Avatar = ""
		err := s.orm.WithContext(ctx).Connection(driver.String()).Query().Create(&user)
		s.Nil(err)
		s.Equal("with_context_name", user.Name)
		s.Equal("with_context_goravel", user.Avatar)
	}
}

type UserObserver struct{}

func (u *UserObserver) Retrieved(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) Creating(event contractsorm.Event) error {
	name := event.GetAttribute("name")
	if name != nil {
		if name.(string) == "observer_name" {
			return errors.New("error")
		}
		if name.(string) == "with_context_name" {
			if avatar := event.Context().Value(testContextKey); avatar != nil {
				event.SetAttribute("avatar", avatar.(string))
			}
		}
	}

	return nil
}

func (u *UserObserver) Created(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) Updating(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) Updated(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) Saving(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) Saved(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) Deleting(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) Deleted(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) ForceDeleting(event contractsorm.Event) error {
	return nil
}

func (u *UserObserver) ForceDeleted(event contractsorm.Event) error {
	return nil
}
