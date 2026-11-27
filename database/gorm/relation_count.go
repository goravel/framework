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
	if len(callback) > 1 {
		return r.withError(errors.OrmQueryInvalidParameter)
	}

	relation, column, aliased := cutRelationAlias(relation)
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
		sub, err := r.relationSubquery(parent, table, count.relation, count.callback)
		if err != nil {
			r.conditions.err = errors.Join(r.conditions.err, err)
			continue
		}

		if target != nil && target.LookUpField(count.column) == nil {
			r.conditions.err = errors.Join(r.conditions.err, errors.OrmQueryRelationCountNoField.Args(target.Name, count.column))
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
		selected = strings.Join(r.relationCountSelectedColumns(parent), ", ")
		r.conditions.selectColumns = nil
	case r.conditions.selectRaw != nil:
		raw, ok := r.conditions.selectRaw.query.(string)
		if !ok {
			r.conditions.err = errors.Join(r.conditions.err, errors.OrmQueryInvalidParameter)
			return db
		}
		selected = raw
		args = append(r.conditions.selectRaw.args, args...)
	case len(db.Statement.Omits) > 0:
		// A raw select skips GORM's Omit, so the model's columns are listed
		// without the omitted ones.
		selected = strings.Join(r.relationCountColumnsWithoutOmits(parent, table, db.Statement.Omits), ", ")
	}

	r.conditions.selectRaw = &Select{
		query: selected + ", " + strings.Join(columns, ", "),
		args:  args,
	}

	return db
}

// relationCountSelectedColumns returns the selected columns, with a field
// name resolved to its quoted column as GORM's Select does. Other
// expressions are kept as they are.
func (r *Query) relationCountSelectedColumns(parent *schema.Schema) []string {
	columns := make([]string, 0, len(r.conditions.selectColumns))
	for _, column := range r.conditions.selectColumns {
		if field := parent.LookUpField(strings.TrimSpace(column)); field != nil && field.DBName != "" {
			column = r.quote(field.DBName)
		}
		columns = append(columns, column)
	}

	return columns
}

// relationCountColumnsWithoutOmits returns the model's columns, qualified
// with the table, without the omitted ones. Read-only fields, such as the
// `gorm:"->"` fields of the counts, are left out too, because they usually
// aren't columns of the table.
func (r *Query) relationCountColumnsWithoutOmits(parent *schema.Schema, table string, omits []string) []string {
	omitted := make(map[string]bool, len(omits))
	for _, omit := range omits {
		if field := parent.LookUpField(omit); field != nil {
			omit = field.DBName
		}
		omitted[omit] = true
	}

	var columns []string
	for _, field := range parent.Fields {
		if field.DBName == "" || omitted[field.DBName] || (!field.Creatable && !field.Updatable) {
			continue
		}
		columns = append(columns, r.quote(table+"."+field.DBName))
	}

	return columns
}

// relationCountTarget returns the schema of the struct the rows are read
// into, which needs a field for each count, or nil when they are read into
// maps. A query without a destination, such as Cursor, reads into maps.
func (r *Query) relationCountTarget(db *gormio.DB, parent *schema.Schema) (*schema.Schema, error) {
	if r.conditions.dest == nil {
		return nil, nil
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

// cutRelationAlias splits "Books as published_count" into the relation and
// the column name.
func cutRelationAlias(relation string) (string, string, bool) {
	fields := strings.Fields(relation)
	if len(fields) == 3 && strings.EqualFold(fields[1], "as") {
		return fields[0], fields[2], true
	}

	return strings.TrimSpace(relation), "", false
}
