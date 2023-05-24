package gorm

import (
	"gorm.io/gorm"

	"github.com/goravel/framework/contracts/database/orm"
)

type Transaction struct {
	orm.Query
	instance *gorm.DB
}

func NewTransaction(tx *gorm.DB) *Transaction {
	return &Transaction{Query: NewQueryWithWithoutEvents(tx, false), instance: tx}
}

func (r *Transaction) Commit() error {
	return r.instance.Commit().Error
}

func (r *Transaction) Rollback() error {
	return r.instance.Rollback().Error
}
