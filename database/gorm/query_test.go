package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gormio "gorm.io/gorm"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

func TestGetObserver(t *testing.T) {
	query := &Query{
		modelToObserver: []contractsorm.ModelToObserver{
			{
				Model:    User{},
				Observer: &UserObserver{},
			},
		},
	}

	assert.Nil(t, query.getObserver(Product{}))
	assert.Equal(t, &UserObserver{}, query.getObserver(User{}))
}

func TestFilterFindConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions []any
		expectErr  error
	}{
		{
			name: "condition is empty",
		},
		{
			name:       "condition is empty string",
			conditions: []any{""},
			expectErr:  errors.OrmMissingWhereClause,
		},
		{
			name:       "condition is empty slice",
			conditions: []any{[]string{}},
			expectErr:  errors.OrmMissingWhereClause,
		},
		{
			name:       "condition has value",
			conditions: []any{"name = ?", "test"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := filterFindConditions(test.conditions...)
			if test.expectErr != nil {
				assert.Equal(t, err, test.expectErr)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetDeletedAtColumnName(t *testing.T) {
	type Test1 struct {
		Deleted gormio.DeletedAt
	}

	assert.Equal(t, "Deleted", getDeletedAtColumn(Test1{}))
	assert.Equal(t, "Deleted", getDeletedAtColumn(&Test1{}))

	type Test2 struct {
		Test1
	}

	assert.Equal(t, "Deleted", getDeletedAtColumn(Test2{}))
	assert.Equal(t, "Deleted", getDeletedAtColumn(&Test2{}))
}

func TestGetModelConnection(t *testing.T) {
	tests := []struct {
		name             string
		model            any
		expectConnection string
	}{
		{
			name: "invalid model",
			model: func() any {
				var product string
				return product
			}(),
		},
		{
			name: "not ConnectionModel",
			model: func() any {
				var user User
				return user
			}(),
		},
		{
			name: "the connection of model is empty",
			model: func() any {
				var review Review
				return review
			}(),
		},
		{
			name: "model is map",
			model: func() any {
				return map[string]any{}
			}(),
		},
		{
			name: "the connection of model is not empty",
			model: func() any {
				var product Product
				return product
			}(),
			expectConnection: "sqlite",
		},
		{
			name: "the connection of model is not empty and model is slice",
			model: func() any {
				var products []Product
				return products
			}(),
			expectConnection: "sqlite",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := &Query{
				conditions: Conditions{
					model: test.model,
				},
			}
			connection := query.getModelConnection()

			assert.Equal(t, test.expectConnection, connection)
		})
	}
}

func TestObserverEvent(t *testing.T) {
	assert.EqualError(t, getObserverEvent(contractsorm.EventRetrieved, &UserObserver{})(nil), "retrieved")
	assert.EqualError(t, getObserverEvent(contractsorm.EventCreating, &UserObserver{})(nil), "creating")
	assert.EqualError(t, getObserverEvent(contractsorm.EventCreated, &UserObserver{})(nil), "created")
	assert.EqualError(t, getObserverEvent(contractsorm.EventUpdating, &UserObserver{})(nil), "updating")
	assert.EqualError(t, getObserverEvent(contractsorm.EventUpdated, &UserObserver{})(nil), "updated")
	assert.EqualError(t, getObserverEvent(contractsorm.EventSaving, &UserObserver{})(nil), "saving")
	assert.EqualError(t, getObserverEvent(contractsorm.EventSaved, &UserObserver{})(nil), "saved")
	assert.EqualError(t, getObserverEvent(contractsorm.EventDeleting, &UserObserver{})(nil), "deleting")
	assert.EqualError(t, getObserverEvent(contractsorm.EventDeleted, &UserObserver{})(nil), "deleted")
	assert.EqualError(t, getObserverEvent(contractsorm.EventForceDeleting, &UserObserver{})(nil), "forceDeleting")
	assert.EqualError(t, getObserverEvent(contractsorm.EventForceDeleted, &UserObserver{})(nil), "forceDeleted")
	assert.Nil(t, getObserverEvent("error", &UserObserver{}))
}

type User struct {
	Name string
}

type UserObserver struct{}

func (u *UserObserver) Retrieved(event contractsorm.Event) error {
	return errors.New("retrieved")
}

func (u *UserObserver) Creating(event contractsorm.Event) error {
	return errors.New("creating")
}

func (u *UserObserver) Created(event contractsorm.Event) error {
	return errors.New("created")
}

func (u *UserObserver) Updating(event contractsorm.Event) error {
	return errors.New("updating")
}

func (u *UserObserver) Updated(event contractsorm.Event) error {
	return errors.New("updated")
}

func (u *UserObserver) Saving(event contractsorm.Event) error {
	return errors.New("saving")
}

func (u *UserObserver) Saved(event contractsorm.Event) error {
	return errors.New("saved")
}

func (u *UserObserver) Deleting(event contractsorm.Event) error {
	return errors.New("deleting")
}

func (u *UserObserver) Deleted(event contractsorm.Event) error {
	return errors.New("deleted")
}

func (u *UserObserver) ForceDeleting(event contractsorm.Event) error {
	return errors.New("forceDeleting")
}

func (u *UserObserver) ForceDeleted(event contractsorm.Event) error {
	return errors.New("forceDeleted")
}

type Product struct {
	Name string
}

func (p *Product) Connection() string {
	return "sqlite"
}

type Review struct {
	Body string
}

func (r *Review) Connection() string {
	return ""
}
