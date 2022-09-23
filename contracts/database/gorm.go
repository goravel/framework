package database

import "gorm.io/gorm"

type Gorm interface {
	Connection(name string) Gorm
	Query() *gorm.DB
}
