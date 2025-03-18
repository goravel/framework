package auth

import (
	"gorm.io/gorm/clause"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
)

type OrmUserProvider struct {
	orm orm.Orm
}

// RetriveById implements auth.UserProvider.
func (o OrmUserProvider) RetriveById(user any, id any) error {
	if err := o.orm.Query().FindOrFail(user, clause.Eq{Column: clause.PrimaryColumn, Value: id}); err != nil {
		return err
	}

	return nil
}

func NewOrmUserProvider(providerName string, orm orm.Orm, config config.Config) (contractsauth.UserProvider, error) {
	return OrmUserProvider{
		orm: orm,
	}, nil
}
