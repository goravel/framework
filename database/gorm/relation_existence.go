package gorm

import (
	"math"
	"reflect"
	"slices"
	"strings"

	gormio "gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/str"
)

// relationOperators are the operators accepted to compare the number of
// related records.
var relationOperators = []string{"=", "!=", "<>", "<", "<=", ">", ">="}

// relationExistence is a where condition on the related records of a
// relationship. It is resolved when the conditions are built, once the model
// of the query is known.
type relationExistence struct {
	callback func(contractsorm.Query) contractsorm.Query
	relation string
	operator string
	count    int64
}

func (r *Query) DoesntHave(relation string) contractsorm.Query {
	return r.addRelationExistence(relation, nil, false, []any{"<", 1})
}

func (r *Query) Has(relation string, args ...any) contractsorm.Query {
	return r.addRelationExistence(relation, nil, false, args)
}

func (r *Query) OrDoesntHave(relation string) contractsorm.Query {
	return r.addRelationExistence(relation, nil, true, []any{"<", 1})
}

func (r *Query) OrHas(relation string, args ...any) contractsorm.Query {
	return r.addRelationExistence(relation, nil, true, args)
}

func (r *Query) OrWhereDoesntHave(relation string, callback func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	return r.addRelationExistence(relation, callback, true, []any{"<", 1})
}

func (r *Query) OrWhereHas(relation string, callback func(contractsorm.Query) contractsorm.Query, args ...any) contractsorm.Query {
	return r.addRelationExistence(relation, callback, true, args)
}

func (r *Query) WhereDoesntHave(relation string, callback func(contractsorm.Query) contractsorm.Query) contractsorm.Query {
	return r.addRelationExistence(relation, callback, false, []any{"<", 1})
}

func (r *Query) WhereHas(relation string, callback func(contractsorm.Query) contractsorm.Query, args ...any) contractsorm.Query {
	return r.addRelationExistence(relation, callback, false, args)
}

func (r *Query) addRelationExistence(relation string, callback func(contractsorm.Query) contractsorm.Query, or bool, args []any) contractsorm.Query {
	if relation == "" {
		return r.withError(errors.OrmQueryEmptyRelation)
	}

	operator, count, err := relationCount(args)
	if err != nil {
		return r.withError(err)
	}

	return r.addWhere(contractsdriver.Where{
		Query: relationExistence{
			callback: callback,
			relation: relation,
			operator: operator,
			count:    count,
		},
		Or: or,
	})
}

// buildRelationExistence resolves the relation on the model of the query and
// returns the condition on its related records.
func (r *Query) buildRelationExistence(db *gormio.DB, existence relationExistence) (clause.Expression, error) {
	model := r.conditions.model
	if model == nil {
		model = r.conditions.dest
	}
	if model == nil {
		return nil, errors.OrmQueryRelationWithoutModel.Args(existence.relation)
	}

	statement := &gormio.Statement{DB: db}
	if err := statement.Parse(model); err != nil {
		if errors.Is(err, schema.ErrUnsupportedDataType) {
			return nil, errors.OrmQueryRelationWithoutModel.Args(existence.relation)
		}

		return nil, err
	}

	table := db.Statement.Table
	if table == "" {
		table = statement.Schema.Table
	}

	return r.relationExistenceExpression(statement.Schema, table, existence)
}

// relationExistenceExpression compares the related records of the first
// relation in the path with the operator and the count. A nested path, such
// as "Books.Author", becomes a condition of the first relation's records on
// the rest of the path, which carries the callback and the count.
func (r *Query) relationExistenceExpression(parent *schema.Schema, parentTable string, existence relationExistence) (clause.Expression, error) {
	name, rest, nested := strings.Cut(existence.relation, ".")

	relation, ok := parent.Relationships.Relations[name]
	if !ok {
		return nil, errors.OrmQueryRelationNotFound.Args(name, parent.Name)
	}

	model := reflect.New(relation.FieldSchema.ModelType).Interface()
	if err := r.checkRelationConnection(name, model); err != nil {
		return nil, err
	}

	sub := r.newRelationQuery(model)

	// A relation on the same table, such as the children of a tree node,
	// queries the related records under an alias. Otherwise the table of
	// the subquery hides the parent's, and the condition would compare each
	// related record with itself.
	relatedTable := relation.FieldSchema.Table
	if relatedTable == parentTable {
		relatedTable = str.Of(name).Snake().String() + "_" + relatedTable
		// The schema's table already carries the connection's prefix, so it
		// is set as is rather than through Table, which would add it again.
		// The alias is quoted, so a mixed case table keeps its case, and set
		// as the statement's table, which Table can't parse from it.
		conditions := sub.conditions
		conditions.table = &Table{
			name:  r.quote(clause.Table{Name: relation.FieldSchema.Table}) + " AS " + r.quote(clause.Table{Name: relatedTable}),
			alias: relatedTable,
		}
		sub = sub.setConditions(conditions)
	}

	operator, count := existence.operator, existence.count
	if nested {
		// As in Laravel, "doesn't have Books.Authors" is "doesn't have a
		// book that has authors": the negation stays on the first relation.
		nestedOperator, nestedCount := operator, count
		operator, count = ">=", 1
		if nestedOperator == "<" && nestedCount == 1 {
			nestedOperator, nestedCount = ">=", 1
			operator, count = "<", 1
		}

		sub = sub.addWhere(contractsdriver.Where{
			Query: relationExistence{
				callback: existence.callback,
				relation: rest,
				operator: nestedOperator,
				count:    nestedCount,
			},
		}).(*Query)
	} else if existence.callback != nil {
		query, ok := existence.callback(sub).(*Query)
		if !ok || query == nil {
			return nil, errors.OrmQueryInvalidSubQuery
		}
		sub = query
	}

	db, err := r.buildRelationConditions(sub, model)
	if err != nil {
		return nil, err
	}

	db = r.constrainToParent(db, relation, relatedTable, parentTable)

	switch {
	case operator == ">=" && count == 1:
		return clause.Expr{SQL: "EXISTS (?)", Vars: []any{db.Select("1")}}, nil
	case operator == "<" && count == 1:
		return clause.Expr{SQL: "NOT EXISTS (?)", Vars: []any{db.Select("1")}}, nil
	default:
		return clause.Expr{SQL: "(?) " + operator + " ?", Vars: []any{db.Select("COUNT(*)"), count}}, nil
	}
}

// buildRelationConditions builds the conditions of the related records: the
// callback's and the scopes', global ones included. The scopes are applied
// now rather than when the query runs, so their conditions are grouped with
// the callback's, and an OR among them can't escape the condition on the
// parent row that is added next.
func (r *Query) buildRelationConditions(sub *Query, model any) (*gormio.DB, error) {
	sub = sub.addGlobalScopes()
	scopes := sub.conditions.scopes
	conditions := sub.conditions
	conditions.scopes = nil
	sub = sub.setConditions(conditions)

	db := sub.buildConditions().instance
	for _, scope := range scopes {
		query := NewQuery(r.ctx, r.config, r.dbConfig, db, r.grammar, r.log, r.modelToObserver, nil, r.telemetryResolver).Model(model)
		scoped, ok := scope(query).(*Query)
		if !ok || scoped == nil {
			return nil, errors.OrmQueryInvalidSubQuery
		}
		db = scoped.buildConditions().instance
	}
	if db.Error != nil {
		return nil, db.Error
	}

	groupWhere(db)

	return db, nil
}

// groupWhere puts the where conditions of the statement in parentheses when
// one of them is joined with OR, so a condition added after them applies to
// all of them.
func groupWhere(db *gormio.DB) {
	where, ok := db.Statement.Clauses["WHERE"]
	if !ok {
		return
	}

	expression, ok := where.Expression.(clause.Where)
	if !ok || len(expression.Exprs) < 2 {
		return
	}

	if !slices.ContainsFunc(expression.Exprs, func(expr clause.Expression) bool {
		_, or := expr.(clause.OrConditions)
		return or
	}) {
		return
	}

	expression.Exprs = []clause.Expression{clause.And(expression.Exprs...)}
	where.Expression = expression
	db.Statement.Clauses["WHERE"] = where
}

// constrainToParent limits the related records to the ones of the parent
// row, joining the pivot table of a many to many relation. The related table
// is the table of the related records in the subquery, or its alias.
func (r *Query) constrainToParent(db *gormio.DB, relation *schema.Relationship, relatedTable, parentTable string) *gormio.DB {
	if relation.JoinTable != nil {
		pivot := relation.JoinTable.Table

		var on []string
		for _, reference := range relation.References {
			switch {
			case reference.OwnPrimaryKey:
				db = db.Where(columnsEqual(pivot, reference.ForeignKey.DBName, parentTable, reference.PrimaryKey.DBName))
			case reference.PrimaryKey != nil:
				on = append(on, r.quote(clause.Column{Table: pivot, Name: reference.ForeignKey.DBName})+" = "+r.quote(clause.Column{Table: relatedTable, Name: reference.PrimaryKey.DBName}))
			default:
				db = db.Where(clause.Eq{Column: clause.Column{Table: pivot, Name: reference.ForeignKey.DBName}, Value: reference.PrimaryValue})
			}
		}

		return db.Joins("INNER JOIN " + r.quote(clause.Table{Name: pivot}) + " ON " + strings.Join(on, " AND "))
	}

	for _, reference := range relation.References {
		switch {
		case reference.PrimaryKey == nil:
			// The type column of a polymorphic relation.
			db = db.Where(clause.Eq{Column: clause.Column{Table: relatedTable, Name: reference.ForeignKey.DBName}, Value: reference.PrimaryValue})
		case reference.OwnPrimaryKey:
			// Has one and has many: the related records point to the parent.
			db = db.Where(columnsEqual(relatedTable, reference.ForeignKey.DBName, parentTable, reference.PrimaryKey.DBName))
		default:
			// Belongs to: the parent points to the related record.
			db = db.Where(columnsEqual(relatedTable, reference.PrimaryKey.DBName, parentTable, reference.ForeignKey.DBName))
		}
	}

	return db
}

// checkRelationConnection returns an error when the related model uses
// another connection than the query: the subquery runs inside the query, so
// it can't reach another database.
func (r *Query) checkRelationConnection(relation string, model any) error {
	connectionModel, ok := model.(contractsorm.ModelWithConnection)
	if !ok {
		return nil
	}

	if connection := connectionModel.Connection(); connection != "" && connection != r.dbConfig.Connection {
		return errors.OrmQueryRelationDifferentConnection.Args(relation, connection, r.dbConfig.Connection)
	}

	return nil
}

// newRelationQuery returns an empty query on the related model, sharing the
// connection of this query.
func (r *Query) newRelationQuery(model any) *Query {
	db := r.instance.Session(&gormio.Session{NewDB: true, Initialized: true})

	return NewQuery(r.ctx, r.config, r.dbConfig, db, r.grammar, r.log, r.modelToObserver, nil, r.telemetryResolver).Model(model).(*Query)
}

func (r *Query) quote(field any) string {
	return r.instance.Statement.Quote(field)
}

func columnsEqual(table, column, otherTable, otherColumn string) clause.Eq {
	return clause.Eq{
		Column: clause.Column{Table: table, Name: column},
		Value:  clause.Column{Table: otherTable, Name: otherColumn},
	}
}

// relationCount reads the optional operator and count of Has and WhereHas,
// which default to at least one related record.
func relationCount(args []any) (string, int64, error) {
	if len(args) == 0 {
		return ">=", 1, nil
	}
	if len(args) != 2 {
		return "", 0, errors.OrmQueryInvalidRelationCount
	}

	operator, ok := args[0].(string)
	if !ok || !slices.Contains(relationOperators, operator) {
		return "", 0, errors.OrmQueryInvalidRelationCount
	}

	var count int64
	value := reflect.ValueOf(args[1])
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		count = value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value.Uint() > math.MaxInt64 {
			return "", 0, errors.OrmQueryInvalidRelationCount
		}
		count = int64(value.Uint())
	default:
		return "", 0, errors.OrmQueryInvalidRelationCount
	}

	if count < 0 {
		return "", 0, errors.OrmQueryInvalidRelationCount
	}

	return operator, count, nil
}
