package auth

import (
	"errors"
	"fmt"
	"reflect"

	contractsauth "github.com/goravel/framework/contracts/auth"
	"github.com/goravel/framework/contracts/config"
	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/foundation"
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
func (o OrmUserProvider) RetriveById(id any) (any, error) {
	user := reflect.New(o.model)

	if err := o.orm.Query().FindOrFail(user, clause.Eq{Column: clause.PrimaryColumn, Value: id}); err != nil {
		return nil, err
	}

	return user, nil
}

func NewOrmUserProvider(name string, orm orm.Orm, config config.Config) (contractsauth.UserProvider, error) {
	model := config.Get(fmt.Sprintf("auth.providers.%s.model", name))

	if model, ok := model.(reflect.Type); ok {
		return OrmUserProvider{
			orm:   orm,
			model: model,
		}, nil
	}

	return nil, errors.New(fmt.Sprintf("You must define the auth.providers.%s.model to create user_provider", name))
}
