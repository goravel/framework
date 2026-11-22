package gorm

import (
	"strings"

	gormio "gorm.io/gorm"
	"gorm.io/gorm/schema"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/deep"
	"github.com/goravel/framework/support/str"
)

// relationCountColumn is a column with the number of the related records of a
// relationship, selected next to the columns of the model.
type relationCountColumn struct {
	callback func(contractsorm.Query) contractsorm.Query
	relation string
	column   string
}

func (r *Query) WithCount(relation string, callback ...func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	relation, column, aliased := cutAlias(relation)
	if relation == "" {
		return r.withError(errors.OrmQueryEmptyRelation)
	}
	if !aliased {
		column = str.Of(relation).Snake().String() + "_count"
	}

	count := relationCountColumn{
		relation: relation,
		column:   column,
	}
	if len(callback) > 0 {
		count.callback = callback[0]
	}

	conditions := r.conditions
	conditions.withCount = deep.Append(conditions.withCount, count)

	return r.setConditions(conditions)
}

// buildWithCount adds the count columns to the selected columns, which are
// the model's own columns when none is selected.
func (r *Query) buildWithCount(db *gormio.DB) *gormio.DB {
	if len(r.conditions.withCount) == 0 {
		return db
	}

	counts := r.conditions.withCount
	r.conditions.withCount = nil

	parent, table, err := r.relationParent(db, counts[0].relation)
	if err != nil {
		r.conditions.err = errors.Join(r.conditions.err, err)
		return db
	}

	target, err := r.relationCountTarget(db, parent)
	if err != nil {
		r.conditions.err = errors.Join(r.conditions.err, err)
		return db
	}

	var (
		columns []string
		args    []any
	)
	for _, count := range counts {
		if target != nil && target.LookUpField(count.column) == nil {
			r.conditions.err = errors.Join(r.conditions.err, errors.OrmQueryRelationCountNoField.Args(target.Name, count.column))
			continue
		}

		sub, err := r.relationSubquery(parent, table, count.relation, count.callback)
		if err != nil {
			r.conditions.err = errors.Join(r.conditions.err, err)
			continue
		}

		columns = append(columns, "(?) AS "+r.quote(count.column))
		args = append(args, sub.Select("COUNT(*)"))
	}
	if len(columns) == 0 {
		return db
	}

	selected := r.quote(table) + ".*"
	switch {
	case len(r.conditions.selectColumns) > 0:
		selected = strings.Join(r.conditions.selectColumns, ", ")
		r.conditions.selectColumns = nil
	case r.conditions.selectRaw != nil:
		raw, ok := r.conditions.selectRaw.query.(string)
		if !ok {
			r.conditions.err = errors.Join(r.conditions.err, errors.OrmQueryInvalidParameter)
			return db
		}
		selected = raw
		args = append(r.conditions.selectRaw.args, args...)
	}

	r.conditions.selectRaw = &Select{
		query: selected + ", " + strings.Join(columns, ", "),
		args:  args,
	}

	return db
}

// relationCountTarget returns the schema of the struct the rows are read
// into, which needs a field for each count, or nil when they are read into
// maps.
func (r *Query) relationCountTarget(db *gormio.DB, parent *schema.Schema) (*schema.Schema, error) {
	if r.conditions.dest == nil {
		return parent, nil
	}

	statement := &gormio.Statement{DB: db}
	if err := statement.Parse(r.conditions.dest); err != nil {
		if errors.Is(err, schema.ErrUnsupportedDataType) {
			return nil, nil
		}

		return nil, err
	}

	return statement.Schema, nil
}

// cutAlias splits "Books as published_count" into the relation and the
// column name.
func cutAlias(relation string) (string, string, bool) {
	fields := strings.Fields(relation)
	if len(fields) == 3 && strings.EqualFold(fields[1], "as") {
		return fields[0], fields[2], true
	}

	return strings.TrimSpace(relation), "", false
}
