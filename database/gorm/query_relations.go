package gorm

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	gormio "gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/deep"
)

// relationshipAliasMaxLength keeps a generated table alias shorter than the identifier limit of every supported
// driver, so PostgreSQL (63) and MySQL (64) accept it.
const relationshipAliasMaxLength = 60

func (r *Query) Has(relationship string, args ...any) contractsorm.Query {
	operator, count, err := hasArguments(relationship, args...)
	if err != nil {
		return r.withError(err)
	}

	return r.addHas(hasCondition{
		relationship: relationship,
		operator:     operator,
		count:        count,
	})
}

func (r *Query) WhereHas(relationship string, callback func(query contractsorm.Query) contractsorm.Query, args ...any) contractsorm.Query {
	operator, count, err := hasArguments(relationship, args...)
	if err != nil {
		return r.withError(err)
	}

	return r.addHas(hasCondition{
		relationship: relationship,
		callback:     callback,
		operator:     operator,
		count:        count,
	})
}

func (r *Query) addHas(condition hasCondition) contractsorm.Query {
	conditions := r.conditions
	conditions.has = deep.Append(conditions.has, condition)

	return r.setConditions(conditions)
}

// buildHas applies the pending relationship filters. A relationship filter is resolved here, not when Has or
// WhereHas is called, because the parent model may only be known from the destination of the terminal operation.
func (r *Query) buildHas(db *gormio.DB) *gormio.DB {
	if len(r.conditions.has) == 0 {
		return db
	}

	has := r.conditions.has
	r.conditions.has = nil
	for _, condition := range has {
		expression, err := r.buildHasCondition(condition)
		if err != nil {
			r.conditions.err = errors.Join(r.conditions.err, err)
			continue
		}

		db = db.Where(expression)
	}

	return db
}

func (r *Query) buildHasCondition(condition hasCondition) (clause.Expression, error) {
	parentSchema, relationship, err := r.resolveRelationship(condition.relationship)
	if err != nil {
		return nil, err
	}

	sub, err := r.buildRelatedQuery(condition, parentSchema, relationship)
	if err != nil {
		return nil, err
	}

	switch {
	case hasUsesExists(condition.operator, condition.count):
		return clause.Expr{SQL: "EXISTS (?)", Vars: []any{sub}}, nil
	case hasUsesNotExists(condition.operator, condition.count):
		return clause.Expr{SQL: "NOT EXISTS (?)", Vars: []any{sub}}, nil
	default:
		return clause.Expr{
			SQL:  "(?) " + condition.operator + " ?",
			Vars: []any{sub.Select("COUNT(*)"), condition.count},
		}, nil
	}
}

// hasUsesExists reports whether the count comparison is exactly equivalent to the existence of a related record.
func hasUsesExists(operator string, count int64) bool {
	switch operator {
	case ">=":
		return count == 1
	case ">", "!=", "<>":
		return count == 0
	default:
		return false
	}
}

// hasUsesNotExists reports whether the count comparison is exactly equivalent to the absence of a related record.
func hasUsesNotExists(operator string, count int64) bool {
	switch operator {
	case "=", "<=":
		return count == 0
	default:
		return false
	}
}

// buildRelatedQuery builds the related query, including its own global scopes and soft deletes, and correlates it
// with the parent query through the association metadata.
func (r *Query) buildRelatedQuery(condition hasCondition, parentSchema *schema.Schema, relationship *schema.Relationship) (*gormio.DB, error) {
	relatedSchema := relationship.FieldSchema
	relatedModel := reflect.New(relatedSchema.ModelType).Interface()

	instance := r.instance.Session(&gormio.Session{NewDB: true, Initialized: true})
	sub := NewQuery(r.ctx, r.config, r.dbConfig, instance, r.grammar, r.log, r.modelToObserver, nil, r.telemetryResolver).
		Model(relatedModel).(*Query)

	if condition.callback != nil {
		callbackQuery := condition.callback(sub)
		var ok bool
		if sub, ok = callbackQuery.(*Query); !ok || sub == nil {
			return nil, errors.OrmQueryInvalidSubQuery
		}
	}

	// A relationship whose related model shares the parent table needs a distinct alias inside the subquery, so the
	// inner and outer tables are unambiguous. Other relationships keep the plain related table.
	alias := ""
	if isSelfRelation(parentSchema, relatedSchema) {
		alias = relationshipAlias(condition.relationship, relatedSchema.Table)
	}

	// A subquery runs inside the parent statement, so it cannot switch to another connection.
	if connection := sub.getModelConnection(); connection != "" && connection != r.dbConfig.Connection {
		return nil, errors.OrmQueryRelationshipUnsupported.Args(condition.relationship, "different database connection")
	}

	built := sub.addGlobalScopes().buildConditions()
	if built.instance.Error != nil {
		return nil, built.instance.Error
	}

	if alias != "" {
		built.instance = aliasRelatedTable(built.instance, relatedSchema.Table, alias)
	}

	correlated := correlateRelationship(built.instance, parentSchema, relationship, alias)

	return correlated, nil
}

// aliasRelatedTable reads the related query from the given table under an alias. The FROM expression is a clause
// expression, so the connection dialect quotes both identifiers, and the statement table is the alias, so clauses
// that target the current table (soft deletes included) use it.
func aliasRelatedTable(db *gormio.DB, table, alias string) *gormio.DB {
	db.Statement.Table = alias
	db.Statement.TableExpr = &clause.Expr{
		SQL:  "? AS ?",
		Vars: []any{clause.Table{Name: table}, clause.Column{Name: alias}},
	}

	return db
}

// correlateRelationship correlates the related query with the parent query through the association metadata.
// HasOne and HasMany match the related foreign key to the parent primary key. BelongsTo matches the related
// primary key to the parent foreign key. Many2Many joins the pivot table and matches both sides there, so the
// parent is never joined to the related table. A self relationship uses the alias of the related table inside the
// subquery and the plain table of the parent outside it.
func correlateRelationship(db *gormio.DB, parentSchema *schema.Schema, relationship *schema.Relationship, alias string) *gormio.DB {
	parentTable := relationshipTable(parentSchema.Table)
	relatedTable := relationshipTable(relationship.FieldSchema.Table)
	if alias != "" {
		relatedTable = alias
	}

	if relationship.Type == schema.Many2Many {
		return correlateMany2Many(db, relationship, parentTable, relatedTable)
	}

	for _, reference := range relationship.References {
		relatedColumn, parentColumn := reference.ForeignKey, reference.PrimaryKey
		// For BelongsTo the parent holds the foreign key and the related model holds the primary key.
		if relationship.Type == schema.BelongsTo {
			relatedColumn, parentColumn = reference.PrimaryKey, reference.ForeignKey
		}

		db = db.Where(clause.Eq{
			Column: clause.Column{Table: relatedTable, Name: relatedColumn.DBName},
			Value:  clause.Column{Table: parentTable, Name: parentColumn.DBName},
		})
	}

	return db
}

// correlateMany2Many joins the pivot table and matches every column of both sides of the association, so a
// composite association never matches on a subset of its columns. The relationship metadata is validated before
// this function runs.
func correlateMany2Many(db *gormio.DB, relationship *schema.Relationship, parentTable, relatedTable string) *gormio.DB {
	own, related := many2ManyReferences(relationship)

	joinTable := relationship.JoinTable.Table

	// Adding a join makes GORM list every related column instead of *. Ask for * explicitly to keep the projection
	// the same as for the other relationship types, unless the callback already chose one.
	if len(db.Statement.Selects) == 0 && len(db.Statement.Omits) == 0 {
		db = db.Select("*")
	}

	join := "INNER JOIN ? ON ? = ?"
	joinArgs := []any{
		clause.Table{Name: joinTable},
		clause.Column{Table: joinTable, Name: related[0].ForeignKey.DBName},
		clause.Column{Table: relatedTable, Name: related[0].PrimaryKey.DBName},
	}
	for _, reference := range related[1:] {
		join += " AND ? = ?"
		joinArgs = append(joinArgs,
			clause.Column{Table: joinTable, Name: reference.ForeignKey.DBName},
			clause.Column{Table: relatedTable, Name: reference.PrimaryKey.DBName},
		)
	}

	db = db.Joins(join, joinArgs...)
	for _, reference := range own {
		db = db.Where(clause.Eq{
			Column: clause.Column{Table: joinTable, Name: reference.ForeignKey.DBName},
			Value:  clause.Column{Table: parentTable, Name: reference.PrimaryKey.DBName},
		})
	}

	return db
}

// many2ManyReferences splits the association metadata into the references of the parent side and the related side.
// validateRelationship guarantees both sides are present.
func many2ManyReferences(relationship *schema.Relationship) ([]*schema.Reference, []*schema.Reference) {
	var own, related []*schema.Reference
	for _, reference := range relationship.References {
		if reference.OwnPrimaryKey {
			own = append(own, reference)
		} else {
			related = append(related, reference)
		}
	}

	return own, related
}

// resolveRelationship finds a supported direct relationship of the parent model: HasOne, HasMany, BelongsTo and
// Many2Many. The keys come from the association metadata, so custom keys are not guessed from the field names.
func (r *Query) resolveRelationship(name string) (*schema.Schema, *schema.Relationship, error) {
	if strings.Contains(name, ".") {
		return nil, nil, errors.OrmQueryRelationshipUnsupported.Args(name, "nested")
	}

	model := r.conditions.model
	if model == nil {
		model = r.conditions.dest
	}
	if model == nil {
		return nil, nil, errors.OrmQueryRelationshipModel.Args(name)
	}

	model, err := modelToStruct(model)
	if err != nil {
		return nil, nil, err
	}

	statement := &gormio.Statement{DB: r.instance}
	if err := statement.Parse(model); err != nil {
		return nil, nil, err
	}

	relationship, err := resolveParsedRelationship(name, statement.Schema)
	if err != nil {
		return nil, nil, err
	}

	return statement.Schema, relationship, nil
}

// resolveParsedRelationship selects a supported direct relationship from parsed association metadata.
func resolveParsedRelationship(name string, parentSchema *schema.Schema) (*schema.Relationship, error) {
	relationship, ok := parentSchema.Relationships.Relations[name]
	if !ok || relationship == nil {
		return nil, errors.OrmQueryRelationshipNotFound.Args(name, parentSchema.Name)
	}

	if relationship.Polymorphic != nil {
		return nil, errors.OrmQueryRelationshipUnsupported.Args(name, "polymorphic")
	}

	if !supportedRelationshipType(relationship.Type) {
		return nil, errors.OrmQueryRelationshipUnsupported.Args(name, string(relationship.Type))
	}

	if err := validateRelationship(name, relationship); err != nil {
		return nil, err
	}

	return relationship, nil
}

// supportedRelationshipType reports whether the relationship type has a correlation strategy.
func supportedRelationshipType(relationshipType schema.RelationshipType) bool {
	switch relationshipType {
	case schema.HasOne, schema.HasMany, schema.BelongsTo, schema.Many2Many:
		return true
	default:
		return false
	}
}

// validateRelationship checks the association metadata needed to build the correlation. GORM normally supplies
// complete metadata; these guards turn malformed custom models into an error instead of a panic.
func validateRelationship(name string, relationship *schema.Relationship) error {
	if relationship.FieldSchema == nil || relationship.Schema == nil || len(relationship.References) == 0 {
		return errors.OrmQueryRelationshipUnsupported.Args(name, "missing association references")
	}

	if relationship.Type == schema.Many2Many && relationship.JoinTable == nil {
		return errors.OrmQueryRelationshipUnsupported.Args(name, "missing association references")
	}

	var own, related bool
	for _, reference := range relationship.References {
		if reference == nil || reference.PrimaryKey == nil || reference.ForeignKey == nil {
			return errors.OrmQueryRelationshipUnsupported.Args(name, "missing association references")
		}

		if reference.OwnPrimaryKey {
			own = true
		} else {
			related = true
		}
	}

	if relationship.Type == schema.Many2Many && (!own || !related) {
		return errors.OrmQueryRelationshipUnsupported.Args(name, "missing association references")
	}

	return nil
}

// isSelfRelation reports whether the related model uses the parent table, so the subquery needs an alias to keep
// the inner and outer tables apart.
func isSelfRelation(parentSchema, relatedSchema *schema.Schema) bool {
	return relationshipTable(parentSchema.Table) == relationshipTable(relatedSchema.Table)
}

// relationshipAlias returns a deterministic alias derived from the relationship name and the related table. The
// alias uses lower-case letters, digits, and underscores, so every supported driver accepts it.
func relationshipAlias(name, table string) string {
	alias := sanitizeIdentifier(name + "_" + table)
	if len(alias) > relationshipAliasMaxLength {
		alias = alias[:relationshipAliasMaxLength]
	}

	return alias
}

// sanitizeIdentifier replaces every character that is not a lower-case letter, a digit, or an underscore.
func sanitizeIdentifier(value string) string {
	var builder strings.Builder
	for _, char := range strings.ToLower(value) {
		switch {
		case char >= 'a' && char <= 'z', char >= '0' && char <= '9', char == '_':
			builder.WriteRune(char)
		default:
			builder.WriteByte('_')
		}
	}

	alias := builder.String()
	if alias == "" {
		return "related"
	}
	if alias[0] >= '0' && alias[0] <= '9' {
		return "t_" + alias
	}

	return alias
}

func relationshipTable(table string) string {
	if index := strings.LastIndex(table, "."); index >= 0 {
		return table[index+1:]
	}

	return table
}

// hasArguments validates the variadic arguments of Has and WhereHas. With no arguments the comparison is
// "at least one related record"; otherwise exactly an operator and a non-negative integer count are required.
func hasArguments(relationship string, args ...any) (string, int64, error) {
	if len(args) == 0 {
		return ">=", 1, nil
	}

	if len(args) != 2 {
		return "", 0, errors.OrmQueryRelationshipArguments.Args(relationship)
	}

	operator, ok := args[0].(string)
	if !ok {
		return "", 0, errors.OrmQueryRelationshipOperator.Args(relationship, fmt.Sprint(args[0]))
	}

	switch operator {
	case "=", "!=", "<>", "<", "<=", ">", ">=":
	default:
		return "", 0, errors.OrmQueryRelationshipOperator.Args(relationship, operator)
	}

	count, ok := hasCount(args[1])
	if !ok {
		return "", 0, errors.OrmQueryRelationshipCount.Args(relationship)
	}

	return operator, count, nil
}

func hasCount(value any) (int64, bool) {
	reflectValue := reflect.ValueOf(value)
	switch reflectValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		count := reflectValue.Int()

		return count, count >= 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		count := reflectValue.Uint()
		if count > math.MaxInt64 {
			return 0, false
		}

		return int64(count), true
	default:
		return 0, false
	}
}
