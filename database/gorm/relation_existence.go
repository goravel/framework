package gorm

import (
	"reflect"
	"slices"
	"strings"

	gormio "gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
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
	parent, table, err := r.relationParent(db, existence.relation)
	if err != nil {
		return nil, err
	}

	return r.relationExistenceExpression(parent, table, existence)
}

// relationParent returns the schema of the model of the query and its table,
// on which the relations are resolved.
func (r *Query) relationParent(db *gormio.DB, relation string) (*schema.Schema, string, error) {
	model := r.conditions.model
	if model == nil {
		model = r.conditions.dest
	}
	if model == nil {
		return nil, "", errors.OrmQueryRelationWithoutModel.Args(relation)
	}

	statement := &gormio.Statement{DB: db}
	if err := statement.Parse(model); err != nil {
		if errors.Is(err, schema.ErrUnsupportedDataType) {
			return nil, "", errors.OrmQueryRelationWithoutModel.Args(relation)
		}

		return nil, "", err
	}

	table := db.Statement.Table
	if table == "" {
		table = statement.Schema.Table
	}

	return statement.Schema, table, nil
}

// relationExistenceExpression compares the related records of the first
// relation in the path with the operator and the count. A nested path, such
// as "Books.Author", becomes a condition of the first relation's records on
// the rest of the path, which carries the callback and the count.
func (r *Query) relationExistenceExpression(parent *schema.Schema, parentTable string, existence relationExistence) (clause.Expression, error) {
	name, rest, nested := strings.Cut(existence.relation, ".")

	callback := existence.callback
	operator, count := existence.operator, existence.count
	if nested {
		callback = func(query contractsorm.Query) contractsorm.Query {
			return query.(*Query).addWhere(contractsdriver.Where{
				Query: relationExistence{
					callback: existence.callback,
					relation: rest,
					operator: existence.operator,
					count:    existence.count,
				},
			})
		}
		operator, count = ">=", 1
	}

	built, err := r.relationSubquery(parent, parentTable, name, callback)
	if err != nil {
		return nil, err
	}

	switch {
	case operator == ">=" && count == 1:
		return clause.Expr{SQL: "EXISTS (?)", Vars: []any{built.Select("1")}}, nil
	case operator == "<" && count == 1:
		return clause.Expr{SQL: "NOT EXISTS (?)", Vars: []any{built.Select("1")}}, nil
	default:
		return clause.Expr{SQL: "(?) " + operator + " ?", Vars: []any{built.Select("COUNT(*)"), count}}, nil
	}
}

// relationSubquery returns the query on the related records of the parent
// row through the relation, limited by the callback.
func (r *Query) relationSubquery(parent *schema.Schema, parentTable, name string, callback func(contractsorm.Query) contractsorm.Query) (*gormio.DB, error) {
	relation, ok := parent.Relationships.Relations[name]
	if !ok {
		return nil, errors.OrmQueryRelationNotFound.Args(name, parent.Name)
	}

	sub := r.newRelationQuery(reflect.New(relation.FieldSchema.ModelType).Interface())
	if callback != nil {
		query, ok := callback(sub).(*Query)
		if !ok || query == nil {
			return nil, errors.OrmQueryInvalidSubQuery
		}
		sub = query
	}

	built := r.constrainToParent(sub, relation, parentTable).addGlobalScopes().buildConditions()
	if err := built.instance.Error; err != nil {
		return nil, err
	}

	return built.instance, nil
}

// constrainToParent limits the related records to the ones of the parent
// row, joining the pivot table of a many to many relation.
func (r *Query) constrainToParent(sub *Query, relation *schema.Relationship, parentTable string) *Query {
	relatedTable := relation.FieldSchema.Table

	if relation.JoinTable != nil {
		pivot := relation.JoinTable.Table

		var on []string
		for _, reference := range relation.References {
			switch {
			case reference.OwnPrimaryKey:
				sub = sub.addWhere(contractsdriver.Where{Query: columnsEqual(pivot, reference.ForeignKey.DBName, parentTable, reference.PrimaryKey.DBName)}).(*Query)
			case reference.PrimaryKey != nil:
				on = append(on, r.quote(clause.Column{Table: pivot, Name: reference.ForeignKey.DBName})+" = "+r.quote(clause.Column{Table: relatedTable, Name: reference.PrimaryKey.DBName}))
			default:
				sub = sub.addWhere(contractsdriver.Where{Query: clause.Eq{Column: clause.Column{Table: pivot, Name: reference.ForeignKey.DBName}, Value: reference.PrimaryValue}}).(*Query)
			}
		}

		return sub.Join("INNER JOIN " + r.quote(clause.Table{Name: pivot}) + " ON " + strings.Join(on, " AND ")).(*Query)
	}

	for _, reference := range relation.References {
		switch {
		case reference.PrimaryKey == nil:
			// The type column of a polymorphic relation.
			sub = sub.addWhere(contractsdriver.Where{Query: clause.Eq{Column: clause.Column{Table: relatedTable, Name: reference.ForeignKey.DBName}, Value: reference.PrimaryValue}}).(*Query)
		case reference.OwnPrimaryKey:
			// Has one and has many: the related records point to the parent.
			sub = sub.addWhere(contractsdriver.Where{Query: columnsEqual(relatedTable, reference.ForeignKey.DBName, parentTable, reference.PrimaryKey.DBName)}).(*Query)
		default:
			// Belongs to: the parent points to the related record.
			sub = sub.addWhere(contractsdriver.Where{Query: columnsEqual(relatedTable, reference.PrimaryKey.DBName, parentTable, reference.ForeignKey.DBName)}).(*Query)
		}
	}

	return sub
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

	count := reflect.ValueOf(args[1])
	switch count.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return operator, count.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return operator, int64(count.Uint()), nil
	default:
		return "", 0, errors.OrmQueryInvalidRelationCount
	}
}
