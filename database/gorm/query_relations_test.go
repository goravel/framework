package gorm

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gormio "gorm.io/gorm"
	"gorm.io/gorm/schema"
	gormtests "gorm.io/gorm/utils/tests"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/errors"
)

func TestHas(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func() contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Model(&HasManyParent{})
	}

	tests := []struct {
		name         string
		query        func() contractsorm.Query
		expectedSql  string
		expectedVars []any
	}{
		{
			name:        "default selects parents with at least one related record",
			query:       func() contractsorm.Query { return newQuery().Has("Items") },
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:        "greater or equal one uses exists",
			query:       func() contractsorm.Query { return newQuery().Has("Items", ">=", 1) },
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:        "greater than zero uses exists",
			query:       func() contractsorm.Query { return newQuery().Has("Items", ">", 0) },
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:        "not equal zero uses exists",
			query:       func() contractsorm.Query { return newQuery().Has("Items", "!=", 0) },
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:        "alternate not equal zero uses exists",
			query:       func() contractsorm.Query { return newQuery().Has("Items", "<>", 0) },
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:        "equal zero uses not exists",
			query:       func() contractsorm.Query { return newQuery().Has("Items", "=", 0) },
			expectedSql: "SELECT * FROM `has_many_parents` WHERE NOT EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:        "less or equal zero uses not exists",
			query:       func() contractsorm.Query { return newQuery().Has("Items", "<=", 0) },
			expectedSql: "SELECT * FROM `has_many_parents` WHERE NOT EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:         "greater than one compares the count",
			query:        func() contractsorm.Query { return newQuery().Has("Items", ">", 1) },
			expectedSql:  "SELECT * FROM `has_many_parents` WHERE (SELECT COUNT(*) FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`) > ?",
			expectedVars: []any{int64(1)},
		},
		{
			name:         "unsigned count compiles",
			query:        func() contractsorm.Query { return newQuery().Has("Items", ">=", uint(2)) },
			expectedSql:  "SELECT * FROM `has_many_parents` WHERE (SELECT COUNT(*) FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`) >= ?",
			expectedVars: []any{int64(2)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := test.query().(*Query).buildConditions()
			statement := query.instance.Find(&[]HasManyParent{}).Statement
			assert.NoError(t, statement.Error)
			assert.Equal(t, test.expectedSql, statement.SQL.String())
			if len(test.expectedVars) == 0 {
				assert.Empty(t, statement.Vars)
			} else {
				assert.Equal(t, test.expectedVars, statement.Vars)
			}
		})
	}
}

func TestHasArguments(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func() contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Model(&HasManyParent{})
	}

	tests := []struct {
		name     string
		args     []any
		expected error
	}{
		{
			name:     "operator without count",
			args:     []any{">"},
			expected: errors.OrmQueryRelationshipArguments.Args("Items"),
		},
		{
			name:     "count without operator",
			args:     []any{2},
			expected: errors.OrmQueryRelationshipArguments.Args("Items"),
		},
		{
			name:     "excess arguments",
			args:     []any{">", 2, 3},
			expected: errors.OrmQueryRelationshipArguments.Args("Items"),
		},
		{
			name:     "negative count",
			args:     []any{">", -1},
			expected: errors.OrmQueryRelationshipCount.Args("Items"),
		},
		{
			name:     "float count",
			args:     []any{">", 1.5},
			expected: errors.OrmQueryRelationshipCount.Args("Items"),
		},
		{
			name:     "string count",
			args:     []any{">", "1"},
			expected: errors.OrmQueryRelationshipCount.Args("Items"),
		},
		{
			name:     "nil count",
			args:     []any{">", nil},
			expected: errors.OrmQueryRelationshipCount.Args("Items"),
		},
		{
			name:     "unsupported operator",
			args:     []any{"==", 1},
			expected: errors.OrmQueryRelationshipOperator.Args("Items", "=="),
		},
		{
			name:     "non string operator",
			args:     []any{1, 1},
			expected: errors.OrmQueryRelationshipOperator.Args("Items", "1"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := newQuery().Has("Items", test.args...).(*Query).buildConditions()

			assert.EqualError(t, query.instance.Error, test.expected.Error())
		})
		t.Run("WhereHas "+test.name, func(t *testing.T) {
			query := newQuery().WhereHas("Items", nil, test.args...).(*Query).buildConditions()

			assert.EqualError(t, query.instance.Error, test.expected.Error())
		})
	}
}

func TestWhereHas(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func() contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Model(&HasManyParent{})
	}

	tests := []struct {
		name         string
		query        func() contractsorm.Query
		expectedSql  string
		expectedVars []any
	}{
		{
			name:        "nil callback behaves like has",
			query:       func() contractsorm.Query { return newQuery().WhereHas("Items", nil) },
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name: "callback conditions stay in the subquery",
			query: func() contractsorm.Query {
				return newQuery().WhereHas("Items", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("name", "book")
				})
			},
			expectedSql:  "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `name` = ? AND `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
			expectedVars: []any{"book"},
		},
		{
			name: "callback count uses not exists",
			query: func() contractsorm.Query {
				return newQuery().WhereHas("Items", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("name", "book")
				}, "=", 0)
			},
			expectedSql:  "SELECT * FROM `has_many_parents` WHERE NOT EXISTS (SELECT * FROM `has_many_items` WHERE `name` = ? AND `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
			expectedVars: []any{"book"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := test.query().(*Query).buildConditions()
			statement := query.instance.Find(&[]HasManyParent{}).Statement
			assert.NoError(t, statement.Error)
			assert.Equal(t, test.expectedSql, statement.SQL.String())
			if len(test.expectedVars) == 0 {
				assert.Empty(t, statement.Vars)
			} else {
				assert.Equal(t, test.expectedVars, statement.Vars)
			}
		})
	}

	t.Run("callback must return a query created by the orm", func(t *testing.T) {
		query := newQuery().WhereHas("Items", func(query contractsorm.Query) contractsorm.Query {
			return nil
		}).(*Query).buildConditions()

		assert.EqualError(t, query.instance.Error, errors.OrmQueryInvalidSubQuery.Error())
	})
}

func TestHasNestedWhereGroups(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	// A where group has no destination of its own, so the explicit model of the owning query must reach the
	// innermost group before its relationship filter builds.
	query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).
		Model(&HasManyParent{}).
		Where(func(outer contractsorm.Query) contractsorm.Query {
			return outer.Where(func(inner contractsorm.Query) contractsorm.Query {
				return inner.Has("Items").Where("name", "book")
			})
		}).(*Query).
		buildConditions()

	statement := query.instance.Find(&[]HasManyParent{}).Statement
	assert.NoError(t, statement.Error)
	// The generated SQL correlates the innermost filter with the parent table, so the model reached the
	// innermost group. GORM flattens nested AND groups, so the two groups do not add parentheses.
	assert.Equal(t, "SELECT * FROM `has_many_parents` WHERE `name` = ? AND EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)", statement.SQL.String())
	assert.Equal(t, []any{"book"}, statement.Vars)

	t.Run("a group that sets its own model keeps it", func(t *testing.T) {
		query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).
			Model(&HasManyItem{}).
			Where(func(group contractsorm.Query) contractsorm.Query {
				return group.Model(&HasManyParent{}).Has("HasManyParent")
			}).(*Query).
			buildConditions()

		// HasManyItem has the HasManyParent relationship and HasManyParent does not, so resolving it against the
		// parent model would succeed and the model of the group would be lost.
		assert.EqualError(t, query.instance.Error, errors.OrmQueryRelationshipNotFound.Args("HasManyParent", "HasManyParent").Error())
	})
}

func TestHasRelationshipTypes(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func(model any) contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Model(model)
	}

	tests := []struct {
		name         string
		query        func() contractsorm.Query
		dest         any
		expectedSql  string
		expectedVars []any
	}{
		{
			name:        "has one matches the related foreign key to the parent primary key",
			query:       func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Profile") },
			dest:        &[]HasManyParent{},
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_profiles` WHERE `has_many_profiles`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:        "has one absence uses not exists",
			query:       func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Profile", "=", 0) },
			dest:        &[]HasManyParent{},
			expectedSql: "SELECT * FROM `has_many_parents` WHERE NOT EXISTS (SELECT * FROM `has_many_profiles` WHERE `has_many_profiles`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:         "has one count compares the count",
			query:        func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Profile", ">", 1) },
			dest:         &[]HasManyParent{},
			expectedSql:  "SELECT * FROM `has_many_parents` WHERE (SELECT COUNT(*) FROM `has_many_profiles` WHERE `has_many_profiles`.`has_many_parent_id` = `has_many_parents`.`id`) > ?",
			expectedVars: []any{int64(1)},
		},
		{
			name:        "belongs to matches the related primary key to the parent foreign key",
			query:       func() contractsorm.Query { return newQuery(&HasManyItem{}).Has("HasManyParent") },
			dest:        &[]HasManyItem{},
			expectedSql: "SELECT * FROM `has_many_items` WHERE EXISTS (SELECT * FROM `has_many_parents` WHERE `has_many_parents`.`id` = `has_many_items`.`has_many_parent_id`)",
		},
		{
			name:        "belongs to absence uses not exists",
			query:       func() contractsorm.Query { return newQuery(&HasManyItem{}).Has("HasManyParent", "=", 0) },
			dest:        &[]HasManyItem{},
			expectedSql: "SELECT * FROM `has_many_items` WHERE NOT EXISTS (SELECT * FROM `has_many_parents` WHERE `has_many_parents`.`id` = `has_many_items`.`has_many_parent_id`)",
		},
		{
			name:        "many to many joins the pivot table",
			query:       func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Roles") },
			dest:        &[]HasManyParent{},
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_roles` INNER JOIN `has_many_parent_role` ON `has_many_parent_role`.`has_many_role_id` = `has_many_roles`.`id` WHERE `has_many_parent_role`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:        "many to many absence uses not exists",
			query:       func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Roles", "=", 0) },
			dest:        &[]HasManyParent{},
			expectedSql: "SELECT * FROM `has_many_parents` WHERE NOT EXISTS (SELECT * FROM `has_many_roles` INNER JOIN `has_many_parent_role` ON `has_many_parent_role`.`has_many_role_id` = `has_many_roles`.`id` WHERE `has_many_parent_role`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
		{
			name:         "many to many count compares the pivot rows",
			query:        func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Roles", ">", 1) },
			dest:         &[]HasManyParent{},
			expectedSql:  "SELECT * FROM `has_many_parents` WHERE (SELECT COUNT(*) FROM `has_many_roles` INNER JOIN `has_many_parent_role` ON `has_many_parent_role`.`has_many_role_id` = `has_many_roles`.`id` WHERE `has_many_parent_role`.`has_many_parent_id` = `has_many_parents`.`id`) > ?",
			expectedVars: []any{int64(1)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := test.query().(*Query).buildConditions()
			statement := query.instance.Find(test.dest).Statement
			assert.NoError(t, statement.Error)
			assert.Equal(t, test.expectedSql, statement.SQL.String())
			if len(test.expectedVars) == 0 {
				assert.Empty(t, statement.Vars)
			} else {
				assert.Equal(t, test.expectedVars, statement.Vars)
			}
		})
	}
}

func TestWhereHasManyToMany(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).
		Model(&HasManyParent{}).
		WhereHas("Roles", func(query contractsorm.Query) contractsorm.Query {
			return query.Where("name", "role")
		}).(*Query).
		buildConditions()

	statement := query.instance.Find(&[]HasManyParent{}).Statement
	assert.NoError(t, statement.Error)
	assert.Equal(t, "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_roles` INNER JOIN `has_many_parent_role` ON `has_many_parent_role`.`has_many_role_id` = `has_many_roles`.`id` WHERE `name` = ? AND `has_many_parent_role`.`has_many_parent_id` = `has_many_parents`.`id`)", statement.SQL.String())
	assert.Equal(t, []any{"role"}, statement.Vars)
}

func TestHasCustomAndCompositeKeys(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func(model any) contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Model(model)
	}

	tests := []struct {
		name        string
		query       func() contractsorm.Query
		dest        any
		expectedSql string
	}{
		{
			name:        "custom foreign key and reference",
			query:       func() contractsorm.Query { return newQuery(&CustomKeyParent{}).Has("Items") },
			dest:        &[]CustomKeyParent{},
			expectedSql: "SELECT * FROM `custom_key_parents` WHERE EXISTS (SELECT * FROM `custom_key_items` WHERE `custom_key_items`.`owner_code` = `custom_key_parents`.`code`)",
		},
		{
			name:        "custom belongs to foreign key and reference",
			query:       func() contractsorm.Query { return newQuery(&CustomKeyItem{}).Has("Parent") },
			dest:        &[]CustomKeyItem{},
			expectedSql: "SELECT * FROM `custom_key_items` WHERE EXISTS (SELECT * FROM `custom_key_parents` WHERE `custom_key_parents`.`code` = `custom_key_items`.`owner_code`)",
		},
		{
			name:        "custom has one foreign key and reference",
			query:       func() contractsorm.Query { return newQuery(&CustomKeyParent{}).Has("Profile") },
			dest:        &[]CustomKeyParent{},
			expectedSql: "SELECT * FROM `custom_key_parents` WHERE EXISTS (SELECT * FROM `custom_key_profiles` WHERE `custom_key_profiles`.`owner_code` = `custom_key_parents`.`code`)",
		},
		{
			name:        "custom has one inverse belongs to foreign key and reference",
			query:       func() contractsorm.Query { return newQuery(&CustomKeyProfile{}).Has("Parent") },
			dest:        &[]CustomKeyProfile{},
			expectedSql: "SELECT * FROM `custom_key_profiles` WHERE EXISTS (SELECT * FROM `custom_key_parents` WHERE `custom_key_parents`.`code` = `custom_key_profiles`.`owner_code`)",
		},
		{
			name:        "every column of a composite reference takes part",
			query:       func() contractsorm.Query { return newQuery(&CompositeParent{}).Has("Items") },
			dest:        &[]CompositeParent{},
			expectedSql: "SELECT * FROM `composite_parents` WHERE EXISTS (SELECT * FROM `composite_items` WHERE `composite_items`.`parent_region` = `composite_parents`.`region` AND `composite_items`.`parent_code` = `composite_parents`.`code`)",
		},
		{
			name:        "composite has one correlates every column",
			query:       func() contractsorm.Query { return newQuery(&CompositeParent{}).Has("FirstItem") },
			dest:        &[]CompositeParent{},
			expectedSql: "SELECT * FROM `composite_parents` WHERE EXISTS (SELECT * FROM `composite_items` WHERE `composite_items`.`parent_region` = `composite_parents`.`region` AND `composite_items`.`parent_code` = `composite_parents`.`code`)",
		},
		{
			name:        "composite belongs to correlates every column in the opposite direction",
			query:       func() contractsorm.Query { return newQuery(&CompositeItem{}).Has("Parent") },
			dest:        &[]CompositeItem{},
			expectedSql: "SELECT * FROM `composite_items` WHERE EXISTS (SELECT * FROM `composite_parents` WHERE `composite_parents`.`region` = `composite_items`.`parent_region` AND `composite_parents`.`code` = `composite_items`.`parent_code`)",
		},
		{
			name:        "composite many to many correlates every own and related reference",
			query:       func() contractsorm.Query { return newQuery(&CompositePost{}).Has("Tags") },
			dest:        &[]CompositePost{},
			expectedSql: "SELECT * FROM `composite_posts` WHERE EXISTS (SELECT * FROM `composite_tags` INNER JOIN `composite_post_tag` ON `composite_post_tag`.`tag_kind` = `composite_tags`.`kind` AND `composite_post_tag`.`tag_code` = `composite_tags`.`code` WHERE `composite_post_tag`.`composite_post_id` = `composite_posts`.`id` AND `composite_post_tag`.`composite_post_tenant_id` = `composite_posts`.`tenant_id`)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := test.query().(*Query).buildConditions()
			statement := query.instance.Find(test.dest).Statement
			assert.NoError(t, statement.Error)
			assert.Equal(t, test.expectedSql, statement.SQL.String())
		})
	}
}

func TestHasSelfRelations(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func(model any) contractsorm.Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).Model(model)
	}

	tests := []struct {
		name        string
		query       func() contractsorm.Query
		dest        any
		expectedSql string
	}{
		{
			name:        "self has many aliases the related table",
			query:       func() contractsorm.Query { return newQuery(&Employee{}).Has("Reports") },
			dest:        &[]Employee{},
			expectedSql: "SELECT * FROM `employees` WHERE EXISTS (SELECT * FROM `employees` AS `reports_employees` WHERE `reports_employees`.`manager_id` = `employees`.`id`)",
		},
		{
			name:        "self has many count compares the aliased count",
			query:       func() contractsorm.Query { return newQuery(&Employee{}).Has("Reports", ">=", 2) },
			dest:        &[]Employee{},
			expectedSql: "SELECT * FROM `employees` WHERE (SELECT COUNT(*) FROM `employees` AS `reports_employees` WHERE `reports_employees`.`manager_id` = `employees`.`id`) >= ?",
		},
		{
			name: "self has many keeps callback conditions in the aliased subquery",
			query: func() contractsorm.Query {
				return newQuery(&Employee{}).WhereHas("Reports", func(query contractsorm.Query) contractsorm.Query {
					return query.Where("name", "report")
				})
			},
			dest:        &[]Employee{},
			expectedSql: "SELECT * FROM `employees` WHERE EXISTS (SELECT * FROM `employees` AS `reports_employees` WHERE `name` = ? AND `reports_employees`.`manager_id` = `employees`.`id`)",
		},
		{
			name:        "self belongs to aliases the related table",
			query:       func() contractsorm.Query { return newQuery(&Employee{}).Has("Manager") },
			dest:        &[]Employee{},
			expectedSql: "SELECT * FROM `employees` WHERE EXISTS (SELECT * FROM `employees` AS `manager_employees` WHERE `manager_employees`.`id` = `employees`.`manager_id`)",
		},
		{
			name:        "self has one aliases the related table and correlates on its foreign key",
			query:       func() contractsorm.Query { return newQuery(&Employee{}).Has("Child") },
			dest:        &[]Employee{},
			expectedSql: "SELECT * FROM `employees` WHERE EXISTS (SELECT * FROM `employees` AS `child_employees` WHERE `child_employees`.`parent_id` = `employees`.`id`)",
		},
		{
			name:        "self many to many aliases the related table and keeps the pivot",
			query:       func() contractsorm.Query { return newQuery(&Employee{}).Has("Friends") },
			dest:        &[]Employee{},
			expectedSql: "SELECT * FROM `employees` WHERE EXISTS (SELECT * FROM `employees` AS `friends_employees` INNER JOIN `employee_friends` ON `employee_friends`.`friend_id` = `friends_employees`.`id` WHERE `employee_friends`.`employee_id` = `employees`.`id`)",
		},
		{
			name:        "non self relation keeps the plain related table",
			query:       func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Items") },
			dest:        &[]HasManyParent{},
			expectedSql: "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := test.query().(*Query).buildConditions()
			statement := query.instance.Find(test.dest).Statement
			assert.NoError(t, statement.Error)
			assert.Equal(t, test.expectedSql, statement.SQL.String())
		})
	}
}

// TestHasSelfRelationWithTablePrefix verifies that the aliased related table is quoted, keeps the table prefix once,
// and that the related soft delete targets the alias.
func TestHasSelfRelationWithTablePrefix(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{
		DryRun:         true,
		NamingStrategy: schema.NamingStrategy{TablePrefix: "tbl_"},
	})
	assert.NoError(t, err)

	query := NewQuery(context.Background(), nil, contractsdatabase.Config{Prefix: "tbl_"}, db, nil, nil, nil, nil, nil).
		Model(&SoftDeleteEmployee{}).
		WhereHas("Reports", nil).(*Query).
		buildConditions()

	statement := query.instance.Find(&[]SoftDeleteEmployee{}).Statement
	assert.NoError(t, statement.Error)

	sql := statement.SQL.String()
	// The prefix appears once in the aliased table, and the dialect quotes both identifiers.
	assert.Contains(t, sql, "FROM `tbl_soft_delete_employees` AS `reports_tbl_soft_delete_employees`")
	assert.NotContains(t, sql, "tbl_tbl_")
	// The related correlation and soft delete use the alias.
	assert.Contains(t, sql, "`reports_tbl_soft_delete_employees`.`manager_id` = `tbl_soft_delete_employees`.`id`")
	assert.Contains(t, sql, "`reports_tbl_soft_delete_employees`.`deleted_at` IS NULL")
}

func TestHasConnection(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func(dbConfig contractsdatabase.Config) *Query {
		return NewQuery(context.Background(), nil, dbConfig, db, nil, nil, nil, nil, nil)
	}

	t.Run("related model without a connection", func(t *testing.T) {
		query := newQuery(contractsdatabase.Config{Connection: "sqlite"}).Model(&HasManyParent{}).Has("Items").(*Query).buildConditions()

		statement := query.instance.Find(&[]HasManyParent{}).Statement
		assert.NoError(t, statement.Error)
		assert.Equal(t, "SELECT * FROM `has_many_parents` WHERE EXISTS (SELECT * FROM `has_many_items` WHERE `has_many_items`.`has_many_parent_id` = `has_many_parents`.`id`)", statement.SQL.String())
	})

	t.Run("related model on the same connection", func(t *testing.T) {
		query := newQuery(contractsdatabase.Config{Connection: "sqlite"}).Model(&SameConnectionParent{}).Has("Items").(*Query).buildConditions()

		statement := query.instance.Find(&[]SameConnectionParent{}).Statement
		assert.NoError(t, statement.Error)
		assert.Equal(t, "SELECT * FROM `same_connection_parents` WHERE EXISTS (SELECT * FROM `same_connection_items` WHERE `same_connection_items`.`same_connection_parent_id` = `same_connection_parents`.`id`)", statement.SQL.String())
	})

	t.Run("related model on a different connection", func(t *testing.T) {
		query := newQuery(contractsdatabase.Config{Connection: "sqlite"}).Model(&OtherConnectionParent{}).Has("Items").(*Query).buildConditions()

		assert.ErrorIs(t, query.instance.Error, errors.OrmQueryRelationshipUnsupported)
		assert.Contains(t, query.instance.Error.Error(), "different database connection")
	})
}

func TestHasErrorPropagation(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func() *Query {
		return NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil)
	}

	t.Run("count returns the missing model error without gorm noise", func(t *testing.T) {
		count, err := newQuery().Has("Items").Count()

		assert.ErrorIs(t, err, errors.OrmQueryRelationshipModel)
		assert.NotContains(t, err.Error(), "Table not set")
		assert.Equal(t, int64(0), count)
	})

	t.Run("exists returns the missing model error without gorm noise", func(t *testing.T) {
		exists, err := newQuery().Has("Items").Exists()

		assert.ErrorIs(t, err, errors.OrmQueryRelationshipModel)
		assert.NotContains(t, err.Error(), "Table not set")
		assert.False(t, exists)
	})

	t.Run("callback build error fails the count", func(t *testing.T) {
		_, err := newQuery().Model(&HasManyParent{}).WhereHas("Items", func(query contractsorm.Query) contractsorm.Query {
			return query.WhereColumn("name", "=", "other", "extra")
		}).Count()

		assert.ErrorIs(t, err, errors.DatabaseInvalidArgumentNumber)
	})
}

func TestHasAggregateErrorPropagation(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func(model any) contractsorm.Query {
		var query contractsorm.Query = NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil)
		if model != nil {
			query = query.Model(model)
		}

		return query
	}

	aggregates := []struct {
		name      string
		aggregate func(query contractsorm.Query) error
	}{
		{
			name: "sum",
			aggregate: func(query contractsorm.Query) error {
				var dest float64
				return query.Sum("id", &dest)
			},
		},
		{
			name: "avg",
			aggregate: func(query contractsorm.Query) error {
				var dest float64
				return query.Avg("id", &dest)
			},
		},
		{
			name: "min",
			aggregate: func(query contractsorm.Query) error {
				var dest float64
				return query.Min("id", &dest)
			},
		},
		{
			name: "max",
			aggregate: func(query contractsorm.Query) error {
				var dest float64
				return query.Max("id", &dest)
			},
		},
	}

	tests := []struct {
		name     string
		query    func() contractsorm.Query
		expected error
	}{
		{
			name:     "missing model",
			query:    func() contractsorm.Query { return newQuery(nil).Has("Items") },
			expected: errors.OrmQueryRelationshipModel,
		},
		{
			name:     "invalid operator",
			query:    func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Items", "==", 1) },
			expected: errors.OrmQueryRelationshipOperator,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, aggregate := range aggregates {
				t.Run(aggregate.name, func(t *testing.T) {
					var aggregateErr error
					require.NotPanics(t, func() {
						aggregateErr = aggregate.aggregate(test.query())
					})

					assert.ErrorIs(t, aggregateErr, test.expected)
				})
			}
		})
	}
}

func TestRelationshipErrors(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	newQuery := func(model any) contractsorm.Query {
		query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil)
		if model != nil {
			query = query.Model(model).(*Query)
		}

		return query
	}

	tests := []struct {
		name     string
		query    func() contractsorm.Query
		expected error
	}{
		{
			name:     "model is not set",
			query:    func() contractsorm.Query { return newQuery(nil).Has("Items") },
			expected: errors.OrmQueryRelationshipModel.Args("Items"),
		},
		{
			name:     "unknown relationship",
			query:    func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Missing") },
			expected: errors.OrmQueryRelationshipNotFound.Args("Missing", "HasManyParent"),
		},
		{
			name:     "invalid model",
			query:    func() contractsorm.Query { return newQuery(map[string]any{}).Has("Items") },
			expected: errors.OrmQueryInvalidModel.Args("map"),
		},
		{
			name:     "nested relationship",
			query:    func() contractsorm.Query { return newQuery(&HasManyParent{}).Has("Items.Name") },
			expected: errors.OrmQueryRelationshipUnsupported.Args("Items.Name", "nested"),
		},
		{
			name:     "polymorphic relationship",
			query:    func() contractsorm.Query { return newQuery(&PolymorphicOwner{}).Has("PolymorphicHouse") },
			expected: errors.OrmQueryRelationshipUnsupported.Args("PolymorphicHouse", "polymorphic"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := test.query().(*Query).buildConditions()

			assert.EqualError(t, query.instance.Error, test.expected.Error())
		})
	}
}

func TestSupportedRelationshipType(t *testing.T) {
	tests := []struct {
		name             string
		relationshipType schema.RelationshipType
		expected         bool
	}{
		{name: "has one", relationshipType: schema.HasOne, expected: true},
		{name: "has many", relationshipType: schema.HasMany, expected: true},
		{name: "belongs to", relationshipType: schema.BelongsTo, expected: true},
		{name: "many to many", relationshipType: schema.Many2Many, expected: true},
		{name: "unsupported type", relationshipType: schema.RelationshipType("unknown"), expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, supportedRelationshipType(test.relationshipType))
		})
	}
}

func TestValidateRelationship(t *testing.T) {
	validSchema := &schema.Schema{Table: "items"}
	validFieldSchema := &schema.Schema{Table: "related_items"}
	primaryKey := &schema.Field{DBName: "id"}
	foreignKey := &schema.Field{DBName: "parent_id"}
	reference := func(ownPrimaryKey bool) *schema.Reference {
		return &schema.Reference{PrimaryKey: primaryKey, ForeignKey: foreignKey, OwnPrimaryKey: ownPrimaryKey}
	}
	relationshipOf := func(relationshipType schema.RelationshipType) *schema.Relationship {
		return &schema.Relationship{
			Type:        relationshipType,
			Schema:      validSchema,
			FieldSchema: validFieldSchema,
			References:  []*schema.Reference{reference(false)},
		}
	}
	missingReferences := errors.OrmQueryRelationshipUnsupported.Args("Items", "missing association references")

	tests := []struct {
		name         string
		relationship *schema.Relationship
		expected     error
	}{
		{
			name:         "nil field schema",
			relationship: &schema.Relationship{Schema: validSchema, References: []*schema.Reference{reference(false)}},
			expected:     missingReferences,
		},
		{
			name:         "nil schema",
			relationship: &schema.Relationship{FieldSchema: validFieldSchema, References: []*schema.Reference{reference(false)}},
			expected:     missingReferences,
		},
		{
			name:         "no references",
			relationship: &schema.Relationship{Schema: validSchema, FieldSchema: validFieldSchema},
			expected:     missingReferences,
		},
		{
			name: "nil reference",
			relationship: &schema.Relationship{
				Type:        schema.HasMany,
				Schema:      validSchema,
				FieldSchema: validFieldSchema,
				References:  []*schema.Reference{nil},
			},
			expected: missingReferences,
		},
		{
			name: "reference without primary key",
			relationship: &schema.Relationship{
				Type:        schema.HasMany,
				Schema:      validSchema,
				FieldSchema: validFieldSchema,
				References:  []*schema.Reference{{ForeignKey: foreignKey}},
			},
			expected: missingReferences,
		},
		{
			name: "reference without foreign key",
			relationship: &schema.Relationship{
				Type:        schema.HasMany,
				Schema:      validSchema,
				FieldSchema: validFieldSchema,
				References:  []*schema.Reference{{PrimaryKey: primaryKey}},
			},
			expected: missingReferences,
		},
		{name: "has many", relationship: relationshipOf(schema.HasMany)},
		{
			name: "many to many without join table",
			relationship: &schema.Relationship{
				Type:        schema.Many2Many,
				Schema:      validSchema,
				FieldSchema: validFieldSchema,
				References:  []*schema.Reference{reference(true), reference(false)},
			},
			expected: missingReferences,
		},
		{
			name: "many to many with only own references",
			relationship: &schema.Relationship{
				Type:        schema.Many2Many,
				Schema:      validSchema,
				FieldSchema: validFieldSchema,
				JoinTable:   &schema.Schema{Table: "pivot"},
				References:  []*schema.Reference{reference(true), reference(true)},
			},
			expected: missingReferences,
		},
		{
			name: "many to many with only related references",
			relationship: &schema.Relationship{
				Type:        schema.Many2Many,
				Schema:      validSchema,
				FieldSchema: validFieldSchema,
				JoinTable:   &schema.Schema{Table: "pivot"},
				References:  []*schema.Reference{reference(false), reference(false)},
			},
			expected: missingReferences,
		},
		{
			name: "many to many",
			relationship: &schema.Relationship{
				Type:        schema.Many2Many,
				Schema:      validSchema,
				FieldSchema: validFieldSchema,
				JoinTable:   &schema.Schema{Table: "pivot"},
				References:  []*schema.Reference{reference(true), reference(false)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRelationship("Items", test.relationship)

			if test.expected == nil {
				assert.NoError(t, err)
				return
			}

			assert.EqualError(t, err, test.expected.Error())
		})
	}
}

func TestResolveParsedRelationship(t *testing.T) {
	validSchema := &schema.Schema{Name: "HasManyParent", Table: "has_many_parents"}
	fieldSchema := &schema.Schema{Table: "has_many_items"}
	primaryKey := &schema.Field{DBName: "id"}
	foreignKey := &schema.Field{DBName: "has_many_parent_id"}
	reference := &schema.Reference{PrimaryKey: primaryKey, ForeignKey: foreignKey}
	valid := &schema.Relationship{
		Name: "Items", Type: schema.HasMany, Schema: validSchema, FieldSchema: fieldSchema,
		References: []*schema.Reference{reference},
	}
	polymorphic := &schema.Relationship{
		Name: "Items", Type: schema.HasOne, Schema: validSchema, FieldSchema: fieldSchema,
		Polymorphic: &schema.Polymorphic{}, References: []*schema.Reference{reference},
	}
	unsupported := &schema.Relationship{
		Name: "Items", Type: schema.RelationshipType("unknown"), Schema: validSchema, FieldSchema: fieldSchema,
		References: []*schema.Reference{reference},
	}
	missingReferences := &schema.Relationship{
		Name: "Items", Type: schema.HasMany, Schema: validSchema, FieldSchema: fieldSchema,
	}

	tests := []struct {
		name      string
		relations map[string]*schema.Relationship
		expected  error
	}{
		{
			name:      "relationship not found",
			relations: map[string]*schema.Relationship{},
			expected:  errors.OrmQueryRelationshipNotFound.Args("Items", "HasManyParent"),
		},
		{
			name:      "nil relationship",
			relations: map[string]*schema.Relationship{"Items": nil},
			expected:  errors.OrmQueryRelationshipNotFound.Args("Items", "HasManyParent"),
		},
		{
			name:      "polymorphic relationship",
			relations: map[string]*schema.Relationship{"Items": polymorphic},
			expected:  errors.OrmQueryRelationshipUnsupported.Args("Items", "polymorphic"),
		},
		{
			name:      "unsupported relationship type",
			relations: map[string]*schema.Relationship{"Items": unsupported},
			expected:  errors.OrmQueryRelationshipUnsupported.Args("Items", "unknown"),
		},
		{
			name:      "missing association references",
			relations: map[string]*schema.Relationship{"Items": missingReferences},
			expected:  errors.OrmQueryRelationshipUnsupported.Args("Items", "missing association references"),
		},
		{
			name:      "valid relationship",
			relations: map[string]*schema.Relationship{"Items": valid},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parent := &schema.Schema{
				Name:          "HasManyParent",
				Relationships: schema.Relationships{Relations: test.relations},
			}

			relationship, err := resolveParsedRelationship("Items", parent)

			if test.expected == nil {
				assert.NoError(t, err)
				assert.Same(t, valid, relationship)
				return
			}

			assert.EqualError(t, err, test.expected.Error())
			assert.Nil(t, relationship)
		})
	}
}

func TestStatementParseError(t *testing.T) {
	db, err := gormio.Open(gormtests.DummyDialector{}, &gormio.Config{DryRun: true})
	assert.NoError(t, err)

	query := NewQuery(context.Background(), nil, contractsdatabase.Config{}, db, nil, nil, nil, nil, nil).
		Model(&InvalidParseModel{}).
		Has("Items").(*Query).
		buildConditions()

	assert.Error(t, query.instance.Error)
}

func TestRelationshipAlias(t *testing.T) {
	tests := []struct {
		name         string
		relationship string
		table        string
		expected     string
		expectedLen  int
	}{
		{
			name:         "alias at the maximum length stays unchanged",
			relationship: strings.Repeat("a", 50),
			table:        strings.Repeat("b", 9),
			expected:     strings.Repeat("a", 50) + "_" + strings.Repeat("b", 9),
			expectedLen:  60,
		},
		{
			name:         "alias longer than the maximum length is truncated",
			relationship: strings.Repeat("a", 51),
			table:        strings.Repeat("b", 9),
			expected:     strings.Repeat("a", 51) + "_" + strings.Repeat("b", 8),
			expectedLen:  60,
		},
		{
			name:         "alias shorter than the maximum length stays unchanged",
			relationship: "reports",
			table:        "employees",
			expected:     "reports_employees",
			expectedLen:  17,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			alias := relationshipAlias(test.relationship, test.table)

			assert.Equal(t, test.expected, alias)
			assert.Len(t, alias, test.expectedLen)
		})
	}
}

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "letters digits and underscores are preserved",
			value:    "a1_b",
			expected: "a1_b",
		},
		{
			name:     "letters at the alphabet edges are preserved",
			value:    "az09",
			expected: "az09",
		},
		{
			name:     "a dot becomes an underscore and upper case becomes lower case",
			value:    "Users.Role",
			expected: "users_role",
		},
		{
			name:     "a dash becomes an underscore",
			value:    "a-b",
			expected: "a_b",
		},
		{
			name:     "a leading zero gets the prefix",
			value:    "0abc",
			expected: "t_0abc",
		},
		{
			name:     "a leading nine gets the prefix",
			value:    "9abc",
			expected: "t_9abc",
		},
		{
			name:     "an empty value becomes the fallback",
			value:    "",
			expected: "related",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, sanitizeIdentifier(test.value))
		})
	}
}

func TestRelationshipTable(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		expected string
	}{
		{
			name:     "schema and table returns the table",
			table:    "public.users",
			expected: "users",
		},
		{
			name:     "dot at the first position returns the table",
			table:    ".users",
			expected: "users",
		},
		{
			name:     "plain table stays unchanged",
			table:    "users",
			expected: "users",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, relationshipTable(test.table))
		})
	}
}

func TestHasCount(t *testing.T) {
	countOf := func(value int64) *int64 {
		return &value
	}

	tests := []struct {
		name       string
		value      any
		expectedOk bool
		// A nil count is not checked, because a rejected value has no promised count.
		expectedCount *int64
	}{
		{
			name:          "largest unsigned count is accepted",
			value:         uint64(math.MaxInt64),
			expectedOk:    true,
			expectedCount: countOf(math.MaxInt64),
		},
		{
			name:          "unsigned count beyond the maximum is rejected",
			value:         uint64(math.MaxInt64) + 1,
			expectedOk:    false,
			expectedCount: countOf(0),
		},
		{
			name:          "largest signed count is accepted",
			value:         int64(math.MaxInt64),
			expectedOk:    true,
			expectedCount: countOf(math.MaxInt64),
		},
		{
			name:       "negative signed count is rejected",
			value:      -1,
			expectedOk: false,
		},
		{
			name:       "small negative signed count is rejected",
			value:      int8(-2),
			expectedOk: false,
		},
		{
			name:          "zero signed count is accepted",
			value:         0,
			expectedOk:    true,
			expectedCount: countOf(0),
		},
		{
			name:          "zero unsigned count is accepted",
			value:         uint(0),
			expectedOk:    true,
			expectedCount: countOf(0),
		},
		{
			name:          "small signed count is accepted",
			value:         int8(2),
			expectedOk:    true,
			expectedCount: countOf(2),
		},
		{
			name:          "small unsigned count is accepted",
			value:         uint8(2),
			expectedOk:    true,
			expectedCount: countOf(2),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			count, ok := hasCount(test.value)

			assert.Equal(t, test.expectedOk, ok)
			if test.expectedCount != nil {
				assert.Equal(t, *test.expectedCount, count)
			}
		})
	}
}

type HasManyParent struct {
	ID      uint `gorm:"primaryKey"`
	Items   []*HasManyItem
	Profile *HasManyProfile
	Roles   []*HasManyRole `gorm:"many2many:has_many_parent_role"`
}

type HasManyItem struct {
	ID              uint `gorm:"primaryKey"`
	HasManyParentID uint
	Name            string
	HasManyParent   *HasManyParent
}

type HasManyRole struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type HasManyProfile struct {
	ID              uint `gorm:"primaryKey"`
	HasManyParentID uint
	Name            string
}

type Employee struct {
	ID        uint `gorm:"primaryKey"`
	ParentID  *uint
	ManagerID *uint
	Manager   *Employee
	Reports   []*Employee    `gorm:"foreignKey:ManagerID"`
	Child     *EmployeeChild `gorm:"foreignKey:ParentID"`
	Friends   []*Employee    `gorm:"many2many:employee_friends"`
}

// EmployeeChild is the related side of a has-one self relation. A struct field of the same type is resolved by GORM
// as belongs-to, so the has-one direction needs its own type over the employees table.
type EmployeeChild struct {
	ID       uint `gorm:"primaryKey"`
	ParentID *uint
}

func (r *EmployeeChild) TableName() string {
	return "employees"
}

// SoftDeleteEmployee is a self relation with soft deletes, so the aliased related table excludes trashed rows.
type SoftDeleteEmployee struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	ManagerID *uint
	Reports   []*SoftDeleteEmployee `gorm:"foreignKey:ManagerID"`
	DeletedAt gormio.DeletedAt
}

// CustomKeyParent associates its items through a non conventional foreign key and reference, and its profile
// uses the same keys in the opposite direction.
type CustomKeyParent struct {
	ID      uint `gorm:"primaryKey"`
	Code    string
	Items   []*CustomKeyItem  `gorm:"foreignKey:OwnerCode;references:Code"`
	Profile *CustomKeyProfile `gorm:"foreignKey:OwnerCode;references:Code"`
}

type CustomKeyItem struct {
	ID        uint `gorm:"primaryKey"`
	OwnerCode string
	Name      string
	Parent    *CustomKeyParent `gorm:"foreignKey:OwnerCode;references:Code"`
}

// CustomKeyProfile is the optional has-one side of CustomKeyParent.
type CustomKeyProfile struct {
	ID        uint `gorm:"primaryKey"`
	OwnerCode string
	Name      string
	Parent    *CustomKeyParent `gorm:"foreignKey:OwnerCode;references:Code"`
}

// CompositeParent associates its items through two columns, so a partial match is not an association. FirstItem is
// the has-one view of the same reference pair.
type CompositeParent struct {
	ID        uint `gorm:"primaryKey"`
	Region    string
	Code      string
	Items     []*CompositeItem `gorm:"foreignKey:ParentRegion,ParentCode;references:Region,Code"`
	FirstItem *CompositeItem   `gorm:"foreignKey:ParentRegion,ParentCode;references:Region,Code"`
}

type CompositeItem struct {
	ID           uint `gorm:"primaryKey"`
	ParentRegion string
	ParentCode   string
	Parent       *CompositeParent `gorm:"foreignKey:ParentRegion,ParentCode;references:Region,Code"`
}

// CompositePost associates tags through a pivot table with two references for each side.
type CompositePost struct {
	ID       uint            `gorm:"primaryKey"`
	TenantID uint            `gorm:"primaryKey"`
	Tags     []*CompositeTag `gorm:"many2many:composite_post_tag;references:Kind,Code;joinReferences:TagKind,TagCode"`
}

type CompositeTag struct {
	ID   uint `gorm:"primaryKey"`
	Kind string
	Code string
}

type PolymorphicOwner struct {
	ID               uint              `gorm:"primaryKey"`
	PolymorphicHouse *PolymorphicHouse `gorm:"polymorphic:Owner"`
}

type PolymorphicHouse struct {
	ID        uint `gorm:"primaryKey"`
	OwnerID   uint
	OwnerType string
}

// SameConnectionParent and SameConnectionItem use the connection of the query, so the related subquery is supported.
type SameConnectionParent struct {
	ID    uint `gorm:"primaryKey"`
	Items []*SameConnectionItem
}

func (r *SameConnectionParent) Connection() string {
	return "sqlite"
}

type SameConnectionItem struct {
	ID                     uint `gorm:"primaryKey"`
	SameConnectionParentID uint
}

func (r *SameConnectionItem) Connection() string {
	return "sqlite"
}

// OtherConnectionParent relates to a model on another connection, which a subquery cannot reach.
type OtherConnectionParent struct {
	ID    uint `gorm:"primaryKey"`
	Items []*OtherConnectionItem
}

func (r *OtherConnectionParent) Connection() string {
	return "sqlite"
}

type OtherConnectionItem struct {
	ID                      uint `gorm:"primaryKey"`
	OtherConnectionParentID uint
}

func (r *OtherConnectionItem) Connection() string {
	return "mysql"
}

// InvalidParseModel has a field GORM cannot map to a column, so statement parsing fails.
type InvalidParseModel struct {
	ID    uint `gorm:"primaryKey"`
	Value map[string]any
}
