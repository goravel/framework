package db

import "gorm.io/gorm"

func Raw(expr string, args ...any) any {
	return gorm.Expr(expr, args...)
}
