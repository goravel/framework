package auth

import (
	"gorm.io/gorm/clause"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/support/database"
)

type OrmUserProvider struct {
	orm orm.Orm
}

func NewOrmUserProvider(providerName string, orm orm.Orm, config config.Config) (contractsauth.UserProvider, error) {
	return &OrmUserProvider{
		orm: orm,
	}, nil
}

// GetID implements auth.UserProvider.
func (r *OrmUserProvider) GetID(user any) any {
	return database.GetID(user)
}

// RetriveByID implements auth.UserProvider.
func (r *OrmUserProvider) RetriveByID(user any, id any) error {
	return r.orm.Query().FindOrFail(user, clause.Eq{Column: clause.PrimaryColumn, Value: id})
}
