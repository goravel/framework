package hints

import (
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WithHint struct {
	Type string
	Keys []string
}

func (indexHint WithHint) ModifyStatement(stmt *gorm.Statement) {
	dialector := sqlserver.Dialector{}
	if stmt.Name() == dialector.Name() {
		for _, name := range []string{"FROM"} {
			clause := stmt.Clauses[name]

			if clause.AfterExpression == nil {
				clause.AfterExpression = indexHint
			} else {
				clause.AfterExpression = Exprs{clause.AfterExpression, indexHint}
			}

			stmt.Clauses[name] = clause
		}
	}
}

func (indexHint WithHint) Build(builder clause.Builder) {
	if len(indexHint.Keys) > 0 {
		_, _ = builder.WriteString(indexHint.Type)
		_ = builder.WriteByte('(')
		for idx, key := range indexHint.Keys {
			if idx > 0 {
				_ = builder.WriteByte(',')
			}
			_, _ = builder.WriteString(key)
		}
		_ = builder.WriteByte(')')
	}
}

func With(names ...string) WithHint {
	return WithHint{Type: "with ", Keys: names}
}

type Exprs []clause.Expression

func (exprs Exprs) Build(builder clause.Builder) {
	for idx, expr := range exprs {
		if idx > 0 {
			_ = builder.WriteByte(' ')
		}
		expr.Build(builder)
	}
}
