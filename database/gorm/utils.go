package gorm

import (
	"fmt"
	"reflect"

	"gorm.io/gorm/clause"

	"github.com/goravel/framework/errors"
)

// notExpression negates a group of where conditions as a whole: NOT (cond1 AND cond2 ...).
// GORM's own Not() negates each condition separately, which is not equivalent when there
// are several conditions, so it can't be used to implement WhereNot.
type notExpression struct {
	exprs []clause.Expression
}

func (r notExpression) Build(builder clause.Builder) {
	_, _ = builder.WriteString("NOT (")
	clause.Where{Exprs: r.exprs}.Build(builder)
	_ = builder.WriteByte(')')
}

func buildWhereColumn(column1 string, column2 ...string) (string, error) {
	switch len(column2) {
	case 1:
		return fmt.Sprintf("%s = %s", column1, column2[0]), nil
	case 2:
		return fmt.Sprintf("%s %s %s", column1, column2[0], column2[1]), nil
	default:
		return "", errors.DatabaseInvalidArgumentNumber.Args(len(column2), "1 or 2")
	}
}

func copyStruct(dest any) reflect.Value {
	t := reflect.TypeOf(dest)
	v := reflect.ValueOf(dest)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
		v = v.Elem()
	}

	destFields := make([]reflect.StructField, 0)
	for field := range t.Fields() {
		destFields = append(destFields, field)
	}
	copyDestStruct := reflect.StructOf(destFields)

	return v.Convert(copyDestStruct)
}
