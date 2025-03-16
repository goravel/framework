package auth

import (
	"fmt"
	"reflect"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"gorm.io/gorm/clause"
)

type OrmUserProvider struct {
	orm   orm.Orm
	model reflect.Type
}

// RetriveByCredentials implements auth.UserProvider.
func (o OrmUserProvider) RetriveByCredentials(credentials map[string]any) (any, error) {
	query := o.orm.Query()

	for key, value := range credentials {
		query.Where(key, value)
	}

	user := reflect.New(o.model)

	if err := query.FirstOrFail(user); err != nil {
		return nil, err
	}

	return user, nil
}

// RetriveById implements auth.UserProvider.
func (o OrmUserProvider) RetriveById(user any, id any) (any, error) {
	if err := o.orm.Query().FindOrFail(user, clause.Eq{Column: clause.PrimaryColumn, Value: id}); err != nil {
		return nil, err
	}

	return user, nil
}

func NewOrmUserProvider(providerName string, orm orm.Orm, config config.Config) (contractsauth.UserProvider, error) {
	model := config.Get(fmt.Sprintf("auth.providers.%s.model", providerName))

	if model, ok := model.(reflect.Type); ok {
		return OrmUserProvider{
			orm:   orm,
			model: model,
		}, nil
	}

	return nil, fmt.Errorf("You must define the auth.providers.%s.model to create user_provider", providerName)
}
