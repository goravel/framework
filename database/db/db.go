package db

import "gorm.io/gorm"

func Raw(expr string, args ...interface{}) any {
	return gorm.Expr(expr, args...)
}
