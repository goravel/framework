package gorm

import (
	"fmt"
	"reflect"

	"gorm.io/gorm/clause"

	"github.com/goravel/framework/database/utils"
	"github.com/goravel/framework/errors"
)

func buildWhereColumn(column1 string, column2 ...string) (string, []any, error) {
	var operator, second string
	switch len(column2) {
	case 1:
		operator, second = "=", column2[0]
	case 2:
		operator, second = column2[0], column2[1]
	default:
		return "", nil, errors.DatabaseInvalidArgumentNumber.Args(len(column2), "1 or 2")
	}

	firstQuery, firstArgs := columnCondition(column1)
	secondQuery, secondArgs := columnCondition(second)

	return fmt.Sprintf("%s %s %s", firstQuery, operator, secondQuery), append(firstArgs, secondArgs...), nil
}

// columnCondition returns the SQL for a column in a condition and the arguments it binds.
// A plain identifier becomes a placeholder bound to a clause.Column, so gorm quotes it when the query is built,
// with the dialect of the connection that runs it. Anything else is an expression and is used as is.
func columnCondition(column string) (string, []any) {
	if !utils.IsPlainIdentifier(column) {
		return column, nil
	}

	return "?", []any{clause.Column{Name: column}}
}

// notExpression negates a group of where conditions as a whole: NOT (cond1 AND cond2 ...).
// GORM's own Not() negates each condition separately, which is not equivalent when there
// are several conditions, so it can't be used to implement WhereNot.
type notExpression struct {
	exprs []clause.Expression
}

func (n notExpression) Build(builder clause.Builder) {
	_, _ = builder.WriteString("NOT (")
	clause.Where{Exprs: n.exprs}.Build(builder)
	_ = builder.WriteByte(')')
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
