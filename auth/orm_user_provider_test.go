package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/errors"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
)

type OrmUserProviderTestSuite struct {
	suite.Suite
	mockContext http.Context
	mockOrm     *mocksorm.Orm
	mockDB      *mocksorm.Query
	provider    *OrmUserProvider
}

func TestOrmUserProviderTestSuite(t *testing.T) {
	suite.Run(t, new(OrmUserProviderTestSuite))
}

func (s *OrmUserProviderTestSuite) SetupTest() {
	s.mockContext = Background()
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.mockDB = mocksorm.NewQuery(s.T())

	ormFacade = s.mockOrm

	provider, err := NewOrmUserProvider(s.mockContext)
	s.Require().Nil(err)
	s.provider = provider.(*OrmUserProvider)
}

func (s *OrmUserProviderTestSuite) TestNewOrmUserProvider_WithNilOrmFacade() {
	ormFacade = nil
	provider, err := NewOrmUserProvider(s.mockContext)
	s.Nil(provider)
	s.ErrorIs(err, errors.OrmFacadeNotSet)
}

func (s *OrmUserProviderTestSuite) TestGetID() {
	type User struct {
		ID uint `gorm:"primaryKey"`
	}
	user := &User{ID: 1}

	id, err := s.provider.GetID(user)
	s.Nil(err)
	s.Equal(uint(1), id)
}

func (s *OrmUserProviderTestSuite) TestRetriveByID() {
	type User struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}
	user := &User{}

	s.mockOrm.EXPECT().WithContext(s.mockContext.Context()).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockDB).Once()
	s.mockDB.EXPECT().FindOrFail(user, mock.Anything).Return(nil).Once()

	err := s.provider.RetriveByID(user, 1)
	s.Nil(err)
}

func (s *OrmUserProviderTestSuite) TestRetriveByID_WithError() {
	type User struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}
	user := &User{}

	s.mockOrm.EXPECT().WithContext(s.mockContext.Context()).Return(s.mockOrm).Once()
	s.mockOrm.EXPECT().Query().Return(s.mockDB).Once()
	s.mockDB.EXPECT().FindOrFail(user, mock.Anything).Return(assert.AnError).Once()

	err := s.provider.RetriveByID(user, 1)
	s.EqualError(err, assert.AnError.Error())
}
